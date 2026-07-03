package modules

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func HTTPServerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, healthCheker *health.HealthCheker, cfg *config.Config, reg *prometheus.Registry) *http.Server {
				mux := http.NewServeMux()
				mux.Handle(cfg.Metrics.Path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

				limiter := rate.NewLimiter(rate.Limit(cfg.App.HealthRateLimit), cfg.App.HealthBurst)
				mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
					if !limiter.Allow() {
						http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
						return
					}

					ctx := r.Context()
					if cfg.App.HealthcheckTimeout > 0 {
						var cancel context.CancelFunc
						ctx, cancel = context.WithTimeout(context.Background(), cfg.App.HealthcheckTimeout)
						defer cancel()
					}

					result := healthCheker.Health(ctx)

					httpStatus := http.StatusOK
					if result.Status != health.HEALTH {
						httpStatus = http.StatusServiceUnavailable
					}

					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(httpStatus)
					err := json.NewEncoder(w).Encode(result)
					if err != nil {
						logger.Error(ctx, "failed encode healthcheck result", zap.Error(err))
					}
				})

				return &http.Server{
					Addr:    net.JoinHostPort(cfg.App.Address, cfg.Metrics.Port),
					Handler: mux,
				}
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, httpServer *http.Server) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							logger.Info("HTTP server started", zap.String("addr", httpServer.Addr))
							if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								logger.Error("HTTP server error", zap.Error(err))
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("stopping HTTP server...")
						return httpServer.Shutdown(ctx)
					},
				})
			},
		),
	)
}
