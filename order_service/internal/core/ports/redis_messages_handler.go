package ports

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type MessageHandler interface {
	Handle(ctx context.Context, msg *redis.Message) error
}
