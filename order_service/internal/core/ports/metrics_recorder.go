package ports

import "context"

type ServiceMetricsRecorder interface {
	OrderCreated(ctx context.Context)
	OrderCompleted(ctx context.Context)
	OrderRejected(ctx context.Context)
	OrderCancelled(ctx context.Context)
	OrderFailedCreate(ctx context.Context)
	OrderFailedUpdate(ctx context.Context)
}
