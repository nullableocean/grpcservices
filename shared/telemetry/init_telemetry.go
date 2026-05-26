package telemetry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ShutdownFunc func(ctx context.Context) error

type Config struct {
	ServiceName      string
	ExporterGRPCAddr string
	RatioSampler     float64

	// если nil, создается стандартный пропагатор
	Propagator propagation.TextMapPropagator

	// если nil, создается стандартный ресурс с service.name = ServiceName
	Resource *resource.Resource

	// если nil, создается новое
	GRPCConn *grpc.ClientConn

	// если nil, используется RatioSampler
	Sampler sdktrace.Sampler

	BatchTimeout       time.Duration
	MaxExportBatchSize int
	MaxQueueSize       int
}

func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return errors.New("invalid telemetry config: ServiceName is required")
	}
	if c.ExporterGRPCAddr == "" {
		return errors.New("invalid telemetry config: ExporterGRPCAddr is required")
	}
	if c.RatioSampler < 0 || c.RatioSampler > 1 {
		return fmt.Errorf("invalid telemetry config: RatioSampler must be between 0 and 1, got %f", c.RatioSampler)
	}

	return nil
}

func NewStandardPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// NewOTLPGrpcExporter создаёт OTLP gRPC экспортёр и gRPC-соединение, если не передано.
func NewOTLPGrpcExporter(ctx context.Context, addr string, conn *grpc.ClientConn) (sdktrace.SpanExporter, func() error, error) {
	connCloser := func() error { return nil }
	if conn == nil {
		var err error
		conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create gRPC client for otlp exporter: %w", err)
		}

		closeOnce := sync.Once{}
		connCloser = func() error {
			var err error

			closeOnce.Do(func() {
				err = conn.Close()
			})

			return err
		}
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		_ = connCloser()
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	return exporter, connCloser, nil
}

func NewTracerProvider(ctx context.Context, cfg *Config) (*sdktrace.TracerProvider, ShutdownFunc, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, nil, err
	}

	var res *resource.Resource
	if cfg.Resource != nil {
		res = cfg.Resource
	} else {
		res, err = newResource(ctx, cfg.ServiceName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed temetry init: error resource creation: %w", err)
		}
	}

	exporter, connCloser, err := NewOTLPGrpcExporter(ctx, cfg.ExporterGRPCAddr, cfg.GRPCConn)
	if err != nil {
		return nil, nil, err
	}

	batchOption := func(batchOpts *sdktrace.BatchSpanProcessorOptions) {
		if cfg.BatchTimeout > 0 {
			batchOpts.BatchTimeout = cfg.BatchTimeout
		}

		if cfg.MaxExportBatchSize > 0 {
			batchOpts.MaxExportBatchSize = cfg.MaxExportBatchSize
		}

		if cfg.MaxQueueSize > 0 {
			batchOpts.MaxQueueSize = cfg.MaxQueueSize
		}
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter, batchOption),
	}

	if cfg.Sampler != nil {
		opts = append(opts, sdktrace.WithSampler(cfg.Sampler))
	} else {
		opts = append(opts, sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.RatioSampler)))
	}

	tp := sdktrace.NewTracerProvider(opts...)

	shutdown := func(ctx context.Context) error {
		return errors.Join(
			tp.Shutdown(ctx),
			connCloser(),
		)
	}

	return tp, shutdown, nil
}

func newResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
}

// SetupGlobal выполняет глоабльную инициализацию OpenTelemetry
func SetupGlobal(ctx context.Context, cfg *Config) (ShutdownFunc, error) {
	tp, shutdown, err := NewTracerProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	propagator := cfg.Propagator
	if propagator == nil {
		propagator = NewStandardPropagator()
	}

	otel.SetTextMapPropagator(propagator)
	otel.SetTracerProvider(tp)

	return shutdown, nil
}
