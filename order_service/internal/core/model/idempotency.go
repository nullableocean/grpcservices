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
}
