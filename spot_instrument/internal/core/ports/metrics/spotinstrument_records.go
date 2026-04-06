package metrics

import "context"

type SpotInstrumentRecords interface {
	ViewMarkets(ctx context.Context)
	FailedViewMarkets(ctx context.Context)
	FailedFindMarket(ctx context.Context)
}
