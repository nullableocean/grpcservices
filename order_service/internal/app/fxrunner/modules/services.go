package modules

import (
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	redis_cache "github.com/nullableocean/grpcservices/orderservice/internal/adapters/cache/rdb"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/client"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ServicesModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *zap.Logger, redisClient *redis.Client, cfg *config.Config, rdbMetrics *metrics.RedisMetricsRecorder) (*redis_cache.IdempotencyCache, *redis_cache.RateLimitCache) {
				idemCache := redis_cache.NewRedisIdempotencyCache(redisClient, cfg.Cache.TTL, rdbMetrics)
				limitCache := redis_cache.NewRedisRateLimitCache(logger, redisClient, rdbMetrics)

				return idemCache, limitCache
			},
		),
		fx.Provide(
			func(cache *redis_cache.RateLimitCache, cfg *config.Config) *access.RoleLimiter {
				rules := map[model.UserRole]access.RoleLimitRule{
					model.UserRoleGuest:       {MaxRequests: cfg.RolesRateLimit.GuestMaxRequests, Window: cfg.RolesRateLimit.GuestWindow},
					model.UserRoleTrader:      {MaxRequests: cfg.RolesRateLimit.TraderMaxRequests, Window: cfg.RolesRateLimit.TraderWindow},
					model.UserRoleMarketMaker: {MaxRequests: cfg.RolesRateLimit.MarketMakerMaxRequests, Window: cfg.RolesRateLimit.MarketMakerWindow},
					model.UserRoleModer:       {MaxRequests: cfg.RolesRateLimit.ModerMaxRequests, Window: cfg.RolesRateLimit.ModerWindow},
					model.UserRoleAdmin:       {MaxRequests: cfg.RolesRateLimit.AdminMaxRequests, Window: cfg.RolesRateLimit.AdminWindow},
				}

				return access.NewRoleLimiter(cache, rules)
			},
		),
		fx.Provide(
			func() *access.AccessService {
				return access.NewRoleAccessService()
			},
		),
		fx.Provide(
			func(
				logger *zap.Logger,
				repo *postgres.OrderRepository,
				spotClient *client.SpotInstrumentClient,
				accessSvc *access.AccessService,
				metricsRecorder *metrics.OrderMetricsRecorder,
				roleLimiter *access.RoleLimiter,
				idemCache *redis_cache.IdempotencyCache,
			) *order.OrderService {
				return order.NewOrderService(
					logger,
					repo,
					spotClient,
					accessSvc,
					metricsRecorder,
					roleLimiter,
					idemCache,
				)
			},
		),
		fx.Provide(func(logger *zap.Logger, grpcServer *grpc.Server, service *order.OrderService, notifier *updatenotifier.UpdateNotifier) *server.OrderServer {
			orderServer := server.NewOrderServer(logger, service, notifier)
			orderv1.RegisterOrderServer(grpcServer, orderServer)
			return orderServer
		}),
	)
}
