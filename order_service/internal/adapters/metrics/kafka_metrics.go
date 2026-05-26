package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type KafkaMetricsRecorder struct {
	messagesPublished     prometheus.Counter
	messagesPublishFailed prometheus.Counter
	sentToDlq             prometheus.Counter
}

func NewKafkaMetricsRecorder(registry *prometheus.Registry) *KafkaMetricsRecorder {
	return &KafkaMetricsRecorder{
		messagesPublished: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_published_total",
			Help: "Total number of messages published to Kafka",
		}),
		messagesPublishFailed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_publish_failed_total",
			Help: "Total number of failed Kafka publish attempts",
		}),
		sentToDlq: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_sent_to_dlq_total",
			Help: "Total number of messages sent to Kafka DLQ",
		}),
	}
}

func (r *KafkaMetricsRecorder) MessagePublished(ctx context.Context) {
	r.messagesPublished.Inc()
}

func (r *KafkaMetricsRecorder) MessagePublishFailed(ctx context.Context) {
	r.messagesPublishFailed.Inc()
}

func (r *KafkaMetricsRecorder) MessageSendToDlq(ctx context.Context) {
	r.sentToDlq.Inc()
}
