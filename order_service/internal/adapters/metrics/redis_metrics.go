package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type RedisMetricsRecorder struct {
	cacheHits      prometheus.Counter
	cacheMisses    prometheus.Counter
	cacheSets      prometheus.Counter
	cacheGetErrors prometheus.Counter
	cacheSetErrors prometheus.Counter
}

func NewRedisMetricsRecorder(registry *prometheus.Registry) *RedisMetricsRecorder {
	return &RedisMetricsRecorder{
		cacheHits: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of Redis cache hits",
		}),
		cacheMisses: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of Redis cache misses",
		}),
		cacheSets: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "redis_cache_sets_total",
			Help: "Total number of Redis cache set operations",
		}),
		cacheGetErrors: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "redis_cache_get_errors_total",
			Help: "Total number of Redis cache get errors",
		}),
		cacheSetErrors: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "redis_cache_set_errors_total",
			Help: "Total number of Redis cache set errors",
		}),
	}
}

func (r *RedisMetricsRecorder) CacheHit(ctx context.Context) {
	r.cacheHits.Inc()
}

func (r *RedisMetricsRecorder) CacheMiss(ctx context.Context) {
	r.cacheMisses.Inc()
}

func (r *RedisMetricsRecorder) CacheSet(ctx context.Context) {
	r.cacheSets.Inc()
}

func (r *RedisMetricsRecorder) CacheGetError(ctx context.Context) {
	r.cacheGetErrors.Inc()
}

func (r *RedisMetricsRecorder) CacheSetError(ctx context.Context) {
	r.cacheSetErrors.Inc()
}
