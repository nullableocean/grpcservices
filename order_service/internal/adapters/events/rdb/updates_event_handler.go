package rdb

import (
	"context"
	"encoding/json"

	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UpdatesMessagesHandler struct {
	notifier *updatenotifier.UpdateNotifier
	logger   *zap.Logger
}

func NewUpdatesMessageHandler(logger *zap.Logger, updateNotifier *updatenotifier.UpdateNotifier) *UpdatesMessagesHandler {
	return &UpdatesMessagesHandler{
		notifier: updateNotifier,
		logger:   logger,
	}
}

func (h *UpdatesMessagesHandler) Handle(ctx context.Context, msg *redis.Message) error {
	updateEvent := &model.EventOrderUpdated{}

	err := json.Unmarshal([]byte(msg.Payload), updateEvent)
	if err != nil {
		h.logger.Warn("failed redis message unmarshal to update event", zap.Error(err))
		return nil
	}

	return h.notifier.Publish(ctx, updateEvent)
}
