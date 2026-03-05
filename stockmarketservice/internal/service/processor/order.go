package processor

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/limiter"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/validator"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type MarketService interface {
	Buy(ctx context.Context, o *domain.Order) error
	Sell(ctx context.Context, o *domain.Order) error
}

type OrderUpdater interface {
	Pending(ctx context.Context, orderUuid string) error
	Reject(ctx context.Context, orderUuid string) error
	Complete(ctx context.Context, orderUuid string) error
}

type StockmarketProcessor struct {
	market     MarketService
	ordUpdater OrderUpdater
	limiter    *limiter.Limiter

	logger *zap.Logger
}

func NewProcessor(logger *zap.Logger, ms MarketService, oUpdater OrderUpdater, processLimit int) *StockmarketProcessor {
	return &StockmarketProcessor{
		market:     ms,
		ordUpdater: oUpdater,
		limiter:    limiter.New(processLimit),

		logger: logger,
	}
}

func (p *StockmarketProcessor) Process(ctx context.Context, o *domain.Order) error {
	if err := validator.ValidateOrder(o); err != nil {
		return err
	}

	p.logger.Info("start processing order",
		zap.String("order_uuid", o.UUID),
		zap.String("user_uuid", o.UserUuid),
		zap.String("market_uuid", o.MarketUuid),
	)

	p.limiter.Acquire()

	go p.process(ctx, o)

	return nil
}

func (p *StockmarketProcessor) process(ctx context.Context, o *domain.Order) {
	defer p.limiter.Release()
	ctx = context.WithoutCancel(ctx)

	ctx, span := otel.Tracer("stockmarket_order_processor").Start(ctx, "process_order")
	defer span.End()

	p.logger.Info("pending order", zap.String("order_uuid", o.UUID))
	err := p.ordUpdater.Pending(ctx, o.UUID)
	if err != nil {
		p.logger.Warn("failed updating, stop process", zap.Error(err))
		span.AddEvent("failed updating order")
	}

	if o.IsBuy() {
		p.logger.Info("handle buy order", zap.String("order_uuid", o.UUID))
		span.AddEvent("failed processing order")

		err := p.market.Buy(ctx, o)
		if err != nil {
			p.logger.Warn("failed buy", zap.String("order_uuid", o.UUID), zap.Error(err))

			err = p.ordUpdater.Reject(ctx, o.UUID)
			if err != nil {
				p.logger.Warn("failed updating status", zap.Error(err))
				span.AddEvent("failed processing order")
				return
			}

			return
		}

		p.logger.Info("success buy process order", zap.String("order_uuid", o.UUID))
		err = p.ordUpdater.Complete(ctx, o.UUID)
		if err != nil {
			p.logger.Warn("failed updating status", zap.Error(err))
			span.AddEvent("failed processing order")
			return
		}
	}

	if o.IsSell() {
		p.logger.Info("handle sell order", zap.String("order_uuid", o.UUID))

		err := p.market.Sell(ctx, o)
		if err != nil {
			p.logger.Warn("failed sell", zap.String("order_uuid", o.UUID), zap.Error(err))
			span.AddEvent("failed processing order")

			err = p.ordUpdater.Reject(ctx, o.UUID)
			if err != nil {
				p.logger.Warn("failed updating status", zap.Error(err))
				span.AddEvent("failed updating order")
				return
			}

			return
		}

		p.logger.Info("success sell process order", zap.String("order_uuid", o.UUID))

		err = p.ordUpdater.Complete(ctx, o.UUID)
		if err != nil {
			p.logger.Warn("failed updating status", zap.Error(err))
			span.AddEvent("failed updating order")
			return
		}
	}
}
