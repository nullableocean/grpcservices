package server

import (
	"context"
	"errors"

	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/errs"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/processor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StockmarketServer struct {
	stockmarketv1.UnimplementedStockMarketServiceServer

	processor *processor.StockmarketProcessor
	logger    *zap.Logger
}

func NewStockmarketServer(logger *zap.Logger, p *processor.StockmarketProcessor) *StockmarketServer {
	return &StockmarketServer{
		processor: p,
		logger:    logger,
	}
}

func (s *StockmarketServer) ProcessOrder(ctx context.Context, req *stockmarketv1.ProcessOrderRequest) (*stockmarketv1.ProcessOrderResponse, error) {
	ctx, span := otel.Tracer("stockmarket_service").Start(ctx, "process_order")
	defer span.End()

	o := mapProtoProcessOrderRequestToDomain(req)
	s.logger.Info("start order process", zap.String("order_uuid", o.UUID))

	span.SetAttributes(attribute.String("order_uuid", o.UUID))

	err := s.processor.Process(ctx, o)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stockmarketv1.ProcessOrderResponse{}, nil
}
