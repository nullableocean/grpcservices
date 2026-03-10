package order

type OrderStatus int

const (
	ORDER_STATUS_CREATED OrderStatus = iota + 1
	ORDER_STATUS_PENDING
	ORDER_STATUS_COMPLETED
	ORDER_STATUS_REJECTED
)

func (status OrderStatus) IsFinal() bool {
	return len(AllowedTransitions(status)) == 0
}

func (status OrderStatus) String() string {
	switch status {
	case ORDER_STATUS_CREATED:
		return "created"
	case ORDER_STATUS_PENDING:
		return "pending"
	case ORDER_STATUS_COMPLETED:
		return "completed"
	case ORDER_STATUS_REJECTED:
		return "rejected"
	}

	return ""
}

// AllowedTransitions
// доступные переходы для текущего статуса
func AllowedTransitions(current OrderStatus) []OrderStatus {
	switch current {
	case ORDER_STATUS_CREATED:
		return []OrderStatus{ORDER_STATUS_PENDING, ORDER_STATUS_COMPLETED, ORDER_STATUS_REJECTED}
	case ORDER_STATUS_PENDING:
		return []OrderStatus{ORDER_STATUS_COMPLETED, ORDER_STATUS_REJECTED}
	case ORDER_STATUS_COMPLETED:
		return nil
	case ORDER_STATUS_REJECTED:
		return nil

	default:
		return nil
	}
}
