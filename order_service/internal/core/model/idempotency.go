package model

type IdempotencyStatus string

const (
	IdempotencyProcessing IdempotencyStatus = "processing"
	IdempotencyCompleted  IdempotencyStatus = "completed"
	IdempotencyFailed     IdempotencyStatus = "failed"
)

type IdempotencyData struct {
	Status    IdempotencyStatus `json:"status"`
	OrderUUID string            `json:"order_uuid,omitempty"`
	LastError string            `json:"last_error,omitempty"`
}

func (d *IdempotencyData) IsCompleted() bool {
	return d.Status == IdempotencyCompleted
}

func (d *IdempotencyData) IsProcessing() bool {
	return d.Status == IdempotencyProcessing
}

func (d *IdempotencyData) IsFailed() bool {
	return d.Status == IdempotencyFailed
}
