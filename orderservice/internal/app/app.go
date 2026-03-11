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
	"github.com/nullableocean/grpcservices/shared/eventbus"
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

type App struct {
	config *config.Config
	logger *zap.Logger

	grpc struct {
		server *grpc.Server

		userservice    *grpc.ClientConn
		spotinstrument *grpc.ClientConn
		stockmarket    *grpc.ClientConn
	}

	http struct {
		server *http.Server
	}

	prometheus struct {
		reg            *prometheus.Registry
		grpcMetricsSrv *grpc_prometheus.ServerMetrics
		grpcMetricsCl  *grpc_prometheus.ClientMetrics

		serviceMetrics *metrics.OrderServiceMetrics
	}

	kafka struct {
		updatesReader        *kafka.Reader
		marketsUpdatesReader *kafka.Reader
		createdEvWriter      *kafka.Writer
		dlqWriter            *kafka.Writer
	}

	redis struct {
		client *redis.Client
	}

	services struct {
		stockmarketEventListener *listener.UpdateListener
		marketsUpdateListener    *listener.SpotInstrumentUpdateListener
	}
}

func NewApp(config *config.Config, logger *zap.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (app *App) Run() error {
	err := app.setupRedis()
	if err != nil {
		return fmt.Errorf("failed setup redis: %w", err)
	}

	app.setupKafka()
	//telemetry
	collectRatio := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(app.config.App.Name, app.config.Telemetry.JaegerGrpcAddress, collectRatio)
	if err != nil {
		return fmt.Errorf("failed init telemetry jaeger exporter: %w", err)
	}
	defer shutdown(context.Background())

	//metrics
	app.setupMetrics()

	//grpc server
	app.setupGrpcServer()

	//grpc clients
	err = app.setupGrpcClients()
	if err != nil {
		return err
	}

	// services init
	spotClient := spotinstrument.NewSpotClient(app.logger, spotv1.NewSpotInstrumentClient(app.grpc.spotinstrument))
	baseSpotSrvs := spot.NewSpotInstrument(spotClient)

	marketsCache := rdb.NewMarketCache(app.logger, app.redis.client, app.config.Redis.TTL)
	cachedSpotSrvs := spot.NewCachedSpotInstrument(baseSpotSrvs, marketsCache, app.logger)

	userClient := userservice.NewUserClient(app.logger, userv1.NewUserClient(app.grpc.userservice))
	userSrvs := user.NewUserService(userClient)

	// events handlers
	updateStatusStreamer := insideHandler.NewStatusStreamer(app.logger, insideHandler.Option{MaxSendingProcess: 5})
	createdEventWriter := writer.NewCreatedEventWriter(app.logger, app.kafka.createdEvWriter)
	createdAmqpEventHandler := insideHandler.NewAmqpOrderCreatedHandler(app.logger, createdEventWriter)

	eventsBus := eventbus.NewEventBus(app.logger, eventbus.Option{})
	eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_NEW_ORDER_STATUS), updateStatusStreamer)
	eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_CREATED_ORDER), createdAmqpEventHandler)

	//main service
	orderSrvs := order.NewOrderService(
		app.logger,
		ram.NewOrderStore(),
		cachedSpotSrvs,
		userSrvs,
		eventsBus,
		access.NewRoleInspector(),
	)

	if app.grpc.stockmarket != nil {
		stockmarketGrpcClient := stockmarketv1.NewStockMarketServiceClient(app.grpc.stockmarket)
		stockMarketClient := transport.NewStockmarketClient(app.logger, stockmarketGrpcClient)
		stockmarket := stockmarket.NewStockMarketService(app.logger, stockMarketClient)
		createdOrderStockmarketHandler := insideHandler.NewStockmarketCreatedOrderHandler(app.logger, orderSrvs, stockmarket)
		eventsBus.RegisterHandler(context.Background(), string(inside.EVENT_CREATED_ORDER), createdOrderStockmarketHandler)
	}

	stockmarketEventsStore := ram.NewEventStore()

	updatesEventHandler := outsideHandlers.NewUpdateEventHandler(app.logger, orderSrvs, stockmarketEventsStore)
	app.services.stockmarketEventListener = listener.NewUpdateListener(
		app.logger,
		app.kafka.updatesReader,
		app.kafka.dlqWriter,
		updatesEventHandler,
		listener.Option{
			ProcessLimit: app.config.Events.ProcLimit,
			MaxRetries:   app.config.Events.Retries,
		},
	)

	app.services.marketsUpdateListener = listener.NewSpotInstrumentUpdateListener(app.logger, app.kafka.marketsUpdatesReader, marketsCache)

	orderServer := server.NewOrderServer(app.logger, orderSrvs, app.prometheus.serviceMetrics, updateStatusStreamer)
	orderv1.RegisterOrderServer(app.grpc.server, orderServer)

	//listen init
	errChan := make(chan error, 1)

	app.startHttpServer(errChan)
	if err := app.startGrpcServer(errChan); err != nil {
		return err
	}

	cancelListen := app.startEventListeners(errChan)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
	case e := <-errChan:
		err = e
	}

	cancelListen()
	app.grpc.server.GracefulStop()
	app.http.server.Shutdown(context.Background())

	return err
}

func (app *App) startEventListeners(errChan chan<- error) context.CancelFunc {
	cancelCtx, cl := context.WithCancel(context.Background())

	go func() {
		err := app.services.stockmarketEventListener.StartListen(cancelCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			errChan <- fmt.Errorf("failed start stockmarket updates listener: %w", err)
		}
	}()

	go func() {
		err := app.services.marketsUpdateListener.StartListen(cancelCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			errChan <- fmt.Errorf("failed start spotmarkets updates listener: %w", err)
		}
	}()

	return cl
}

func (app *App) startGrpcServer(errChan chan<- error) error {
	lis, err := net.Listen("tcp", app.config.App.Address+":"+app.config.App.Port)
	if err != nil {
		return fmt.Errorf("create listen tcp error: %w", err)
	}

	go func() {
		app.logger.Info("order grpc server started", zap.String("address", app.config.App.Address+":"+app.config.App.Port))

		err = app.grpc.server.Serve(lis)
		if err != nil {
			errChan <- fmt.Errorf("start serve grpc error: %w", err)
		}
	}()

	return nil
}

func (app *App) startHttpServer(errChan chan<- error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(app.prometheus.reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	app.http.server = &http.Server{
		Addr:    ":" + app.config.Metrics.Port,
		Handler: mux,
	}

	go func() {
		app.logger.Info("start listen metrics http", zap.String("address", app.config.App.Address+":"+app.config.Metrics.Port))

		err := app.http.server.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
}

func (app *App) setupGrpcServer() {
	app.grpc.server = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		serverUnaryInterceptors(app.logger, app.prometheus.grpcMetricsSrv),
		serverStreamInterceptors(app.logger, app.prometheus.grpcMetricsSrv),
	)
}

func (app *App) setupGrpcClients() error {
	clientInterceptors := clientsInterceptors(app.logger, app.prometheus.grpcMetricsCl)

	spotGrpcConnect, err := grpc.NewClient(
		app.config.Spot.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)
	if err != nil {
		return fmt.Errorf("failed grpc connect to spot service: %w", err)
	}

	userGrpcConnect, err := grpc.NewClient(
		app.config.User.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)
	if err != nil {
		return fmt.Errorf("failed grpc connect to user service: %w", err)
	}

	app.grpc.spotinstrument = spotGrpcConnect
	app.grpc.userservice = userGrpcConnect

	if app.config.Stockmarket.Endpoint != "" {
		stockmarketGrpcConnect, err := grpc.NewClient(
			app.config.Stockmarket.Endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
			clientInterceptors,
		)
		if err != nil {
			return fmt.Errorf("failed grpc connect to stockmarket service: %w", err)
		}

		app.grpc.stockmarket = stockmarketGrpcConnect
	}

	return nil
}

func (app *App) setupMetrics() {
	app.prometheus.grpcMetricsSrv = grpc_prometheus.NewServerMetrics()
	app.prometheus.grpcMetricsCl = grpc_prometheus.NewClientMetrics()

	app.prometheus.reg = prometheus.NewRegistry()
	app.prometheus.reg.MustRegister(
		app.prometheus.grpcMetricsSrv,
		app.prometheus.grpcMetricsCl,
	)

	app.prometheus.serviceMetrics = metrics.NewOrderMetrics(app.prometheus.reg)
}

func (app *App) setupRedis() error {
	c := redis.NewClient(&redis.Options{
		Addr:     app.config.Redis.Address,
		DB:       app.config.Redis.DB,
		Username: app.config.Redis.Username,
		Password: app.config.Redis.Password,
	})

	err := c.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}

	app.redis.client = c

	return nil
}

func (app *App) setupKafka() {
	app.kafka.updatesReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{app.config.Kafka.Endpoint},
		Topic:          app.config.Kafka.OrderUpdatesTopic,
		GroupID:        app.config.Kafka.GroupID,
		MaxWait:        time.Second * 5,
		CommitInterval: 0, // ручной коммит
		StartOffset:    kafka.FirstOffset,
	})

	app.kafka.marketsUpdatesReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{app.config.Kafka.Endpoint},
		Topic:          app.config.Kafka.MarketsUpdateTopic,
		GroupID:        app.config.Kafka.GroupID,
		MaxWait:        time.Second * 5,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})

	app.kafka.createdEvWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{app.config.Kafka.Endpoint},
		Topic:   app.config.Kafka.OrderCreatedTopic,
	})
	app.kafka.createdEvWriter.AllowAutoTopicCreation = true

	app.kafka.dlqWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{app.config.Kafka.Endpoint},
		Topic:   app.config.Kafka.DLQTopic,
	})
	app.kafka.dlqWriter.AllowAutoTopicCreation = true
}
