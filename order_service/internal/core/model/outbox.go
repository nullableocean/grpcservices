package model

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "PENDING"
	OutboxStatusProcessed  OutboxStatus = "PROCESSED"
	OutboxStatusFailed     OutboxStatus = "FAILED"
	OutboxStatusDeadLetter OutboxStatus = "DEAD_LETTER"
)
