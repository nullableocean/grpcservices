package app

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisSetting struct {
	Addr            string
	Password        string
	DB              int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxRetries      int
	MinBackoffDelay time.Duration
	MaxBackoffDelay time.Duration
}

func (a *App) createRedisClient(settings *redisSetting) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            settings.Addr,
		Password:        settings.Password,
		DB:              settings.DB,
		DialTimeout:     settings.DialTimeout,
		ReadTimeout:     settings.ReadTimeout,
		WriteTimeout:    settings.WriteTimeout,
		MaxRetries:      settings.MaxRetries,
		MinRetryBackoff: settings.MinBackoffDelay,
		MaxRetryBackoff: settings.MaxBackoffDelay,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed redis ping: %w", err)
	}

	return client, nil
}
