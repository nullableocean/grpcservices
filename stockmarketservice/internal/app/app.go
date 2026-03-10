package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/config"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/event/order/updater"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/market"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/processor"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/transport/amqp/listener"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/transport/amqp/writer"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/transport/grpc/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Start(cnf *config.Config, logger *zap.Logger) error {
	// telemetry
	ratioTracing := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(cnf.App.Name, cnf.Telemetry.JaegerGrpcAddress, ratioTracing)
	if err != nil {
		return fmt.Errorf("telemtry init error: %w", err)
	}
	defer shutdown(context.Background())

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cnf.Kafka.Endpoint},
		Topic:   cnf.Kafka.OrderUpdatesTopic,
	})
	kafkaWriter.AllowAutoTopicCreation = true

	kfkDlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cnf.Kafka.Endpoint},
		Topic:   cnf.Kafka.DLQTopic,
	})
	kfkDlqWriter.AllowAutoTopicCreation = true

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cnf.Kafka.Endpoint},
		Topic:          cnf.Kafka.OrderCreatedTopic,
		GroupID:        cnf.Kafka.GroupID,
		MaxWait:        time.Second * 5,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})

	// metrics
	grpcMetrics := grpc_prometheus.NewServerMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(grpcMetrics)

	//grpc init
	intersChain := grpc.ChainUnaryInterceptor(
		intercepter.UnaryServerPanicRecovery(logger),
		intercepter.UnaryServerLogger(logger),
		intercepter.UnaryServerTelemtry(),
		grpcMetrics.UnaryServerInterceptor(),
	)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), intersChain)

	// service

	updateWriter := writer.NewOrderUpdateWriter(logger, kafkaWriter)
	updater := updater.NewOrderUpdater(updateWriter)

	dummyMarketService := market.NewMarketService()
	stockProc := processor.NewProcessor(logger, dummyMarketService, updater, cnf.Processing.ProcessLimit)
	stockServer := server.NewStockmarketServer(logger, stockProc)
	stockmarketv1.RegisterStockMarketServiceServer(grpcServer, stockServer)

	createOrderListener := listener.NewCreatedOrderListener(logger, kafkaReader, kfkDlqWriter, stockProc, listener.Option{})
	// server init listen

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(promReg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpServer := &http.Server{
		Addr:    ":" + cnf.Metrics.Port,
		Handler: mux,
	}

	return upAndWaitShutdown(logger, cnf, grpcServer, httpServer, createOrderListener)
}

func upAndWaitShutdown(logger *zap.Logger, cnf *config.Config, grpcServer *grpc.Server, httpServer *http.Server, eventListener *listener.CreatedOrderListener) error {
	var err error
	errChan := make(chan error, 1)

	go func() {
		logger.Info("start listen metrics http", zap.String("address", cnf.App.Address+":"+cnf.Metrics.Port))

		err := httpServer.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	lis, err := net.Listen("tcp", cnf.App.Address+":"+cnf.App.Port)
	if err != nil {
		return fmt.Errorf("create listen tcp error: %w", err)
	}

	go func() {
		logger.Info("stockmarket grpc service started", zap.String("address", cnf.App.Address+":"+cnf.App.Port))

		err = grpcServer.Serve(lis)
		if err != nil {
			errChan <- fmt.Errorf("start serve grpc error: %w", err)
		}
	}()

	listenerCtx, cl := context.WithCancel(context.Background())
	defer cl()

	go func() {
		err = eventListener.StartListen(listenerCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			errChan <- fmt.Errorf("start broker listener error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
		grpcServer.GracefulStop()
		err = httpServer.Shutdown(context.Background())
	case e := <-errChan:
		err = e
	}

	return err
}
