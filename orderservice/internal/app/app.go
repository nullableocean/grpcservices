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
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/cache/rdb"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside/bus"
	insideHandler "github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside/handlers"
	outsideHandlers "github.com/nullableocean/grpcservices/orderservice/internal/service/events/outside/handlers"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/spot"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/stockmarket"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/user"
	"github.com/nullableocean/grpcservices/orderservice/internal/store/ram"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/amqp/listener"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/amqp/writer"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/spotinstrument"
	transport "github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/stockmarket"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/userservice"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/server"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(cnf *config.Config, logger *zap.Logger) error {
	redis, err := setupRedis(cnf)
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cnf.Kafka.Endpoint},
		Topic:          cnf.Kafka.OrderUpdatesTopic,
		GroupID:        cnf.Kafka.GroupID,
		MaxWait:        time.Second * 5,
		CommitInterval: 0, // ручной коммит
		StartOffset:    kafka.FirstOffset,
	})

	kafkaCreatedOrderWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cnf.Kafka.Endpoint},
		Topic:   cnf.Kafka.OrderCreatedTopic,
	})
	kafkaCreatedOrderWriter.AllowAutoTopicCreation = true

	kfkDlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cnf.Kafka.Endpoint},
		Topic:   cnf.Kafka.DLQTopic,
	})
	kfkDlqWriter.AllowAutoTopicCreation = true

	//telemetry
	collectRatio := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(cnf.App.Name, cnf.Telemetry.JaegerGrpcAddress, collectRatio)
	if err != nil {
		return fmt.Errorf("failed init telemetry jaeger exporter: %w", err)
	}
	defer shutdown(context.Background())

	//metrics
	serverMetrics := grpc_prometheus.NewServerMetrics()
	clientMetrics := grpc_prometheus.NewClientMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(serverMetrics, clientMetrics)
	orderServerMetrics := metrics.NewOrderMetrics(promReg)

	//grpc server
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), serverInterceptors(logger, serverMetrics))

	//grpc client
	clientInterceptors := clientsInterceptors(logger, clientMetrics)
	spotGrpcConnect, err := grpc.NewClient(
		cnf.Spot.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)

	userGrpcConnetc, err := grpc.NewClient(
		cnf.User.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)
	if err != nil {
		return fmt.Errorf("failed grpc connect to user service: %w", err)
	}

	stockmarketGrpcConnect, err := grpc.NewClient(
		cnf.Stockmarket.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)
	if err != nil {
		return fmt.Errorf("failed grpc connect to stockmarket service: %w", err)
	}

	// services init

	marketsCache := rdb.NewMarketCache(redis, cnf.Redis.TTL)

	spotClient := spotinstrument.NewSpotClient(spotv1.NewSpotInstrumentClient(spotGrpcConnect))
	spotInstrument := spot.NewSpotInstrument(spotClient)

	cachedSpotInstrument := spot.NewCachedSpotInstrument(spotInstrument, marketsCache, logger)

	userClient := userservice.NewUserClient(logger, userv1.NewUserClient(userGrpcConnetc))
	userService := user.NewUserService(userClient)

	roleInspector := access.NewRoleInspector()

	orderStore := ram.NewOrderStore()

	stockmarketGrpcClient := stockmarketv1.NewStockMarketServiceClient(stockmarketGrpcConnect)
	stockMarketClient := transport.NewStockmarketClient(logger, stockmarketGrpcClient)
	stockmarket := stockmarket.NewStockMarketService(logger, stockMarketClient)

	eventsBus := bus.NewEventBus()

	updateStatusStreamer := insideHandler.NewStatusStreamer(logger, insideHandler.Option{MaxSendingProcess: 5})

	createdEventWriter := writer.NewCreatedEventWriter(logger, kafkaCreatedOrderWriter)
	createdAmqpEventHandler := insideHandler.NewAmqpOrderCreatedHandler(logger, createdEventWriter)

	eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_NEW_ORDER_STATUS), updateStatusStreamer)
	eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_CREATED_ORDER), createdAmqpEventHandler)

	orderService := order.NewOrderService(logger, orderStore, cachedSpotInstrument, userService, eventsBus, roleInspector)

	createdOrderStockmarketHandler := insideHandler.NewStockmarketCreatedOrderHandler(logger, orderService, stockmarket)
	eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_CREATED_ORDER), createdOrderStockmarketHandler)

	outsideEventStore := ram.NewEventStore()

	updatesEventHandler := outsideHandlers.NewUpdateEventHandler(logger, orderService, outsideEventStore)
	updatesEventListener := listener.NewUpdateListener(
		logger,
		kafkaReader,
		kfkDlqWriter,
		updatesEventHandler,
		listener.Option{
			ProcessLimit: cnf.Events.ProcLimit,
			MaxRetries:   cnf.Events.Retries,
		},
	)

	orderServer := server.NewOrderServer(logger, orderService, orderServerMetrics, updateStatusStreamer)
	orderv1.RegisterOrderServer(grpcServer, orderServer)

	//listen init
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

	return upAndWaitShutdown(logger, cnf, grpcServer, httpServer, updatesEventListener)
}

func setupRedis(cnf *config.Config) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     cnf.Redis.Address,
		DB:       cnf.Redis.DB,
		Username: cnf.Redis.Username,
		Password: cnf.Redis.Password,
	})

	err := c.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("redis ping error: %w", err)
	}

	return c, nil
}

func serverInterceptors(logger *zap.Logger, serverMetrics *grpc_prometheus.ServerMetrics) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		intercepter.UnaryServerPanicRecovery(logger),
		intercepter.UnaryServerLogger(logger),
		intercepter.UnaryServerTelemtry(),
		serverMetrics.UnaryServerInterceptor(),
	)
}

func clientsInterceptors(logger *zap.Logger, clientMetrics *grpc_prometheus.ClientMetrics) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(
		intercepter.UnaryClientPanicRecovery(),
		intercepter.UnaryClientXReqId(),
		intercepter.UnaryClientXReqIdTelemtry(),
		clientMetrics.UnaryClientInterceptor(),
		intercepter.UnaryClientLogger(logger),
	)
}

// gracefull
func upAndWaitShutdown(
	logger *zap.Logger, cnf *config.Config,
	grpcServer *grpc.Server,
	httpServer *http.Server,
	updateListener *listener.UpdateListener,
) error {
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
		logger.Info("order grpc service started", zap.String("address", cnf.App.Address+":"+cnf.App.Port))

		err = grpcServer.Serve(lis)
		if err != nil {
			errChan <- fmt.Errorf("start serve grpc error: %w", err)
		}
	}()

	listenerCtx, cl := context.WithCancel(context.Background())
	defer cl()

	go func() {
		err = updateListener.StartListen(listenerCtx)
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
	case e := <-errChan:
		err = e
	}

	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())

	return err
}
