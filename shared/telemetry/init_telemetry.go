package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// закрываем grpc и трейсинг
type ShutdownFunc func(ctx context.Context) error

// InitOpenTelemtryGrpcProvider инициализирует глобальный трейсинг-провайдер с экспортом по grpc
//
// serviceName имя сервиса
//
// exporterGrpcAddr grpc адрес коллектора exmp: jaeger:4317, tempo:4317
//
// ratioSampler частотность трейсинга 0 <= ratio <= 1
func InitOpenTelemtryGrpcProvider(serviceName, exporterGrpcAddr string, ratioSampler float64) (ShutdownFunc, error) {
	ctx := context.Background()
	conn, err := grpc.NewClient(
		exporterGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	// grpc exporter
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// описание сервиса
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// пропоганация, передача между сервисами контекста трейсинга (trace_id, метаданные)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(ratioSampler)),
	)

	otel.SetTracerProvider(provider)

	shutdown := func(ctx context.Context) error {
		return errors.Join(provider.Shutdown(ctx), conn.Close())
	}

	return shutdown, nil
}
