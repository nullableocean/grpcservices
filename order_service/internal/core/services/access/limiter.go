package access

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
)

var _ ports.OrderRateLimiter = &RoleLimiter{}

type RoleLimitRule struct {
	MaxRequests int
	Window      time.Duration
}

type LimitRules map[model.UserRole]RoleLimitRule

type RoleLimiter struct {
	counter ports.CacheRateLimitCounter
	rules   LimitRules
}

func NewRoleLimiter(counterCache ports.CacheRateLimitCounter, rules LimitRules) *RoleLimiter {
	return &RoleLimiter{
		counter: counterCache,
		rules:   rules,
	}
}

func (l *RoleLimiter) Check(ctx context.Context, user model.User) error {
	role := user.HighestPriorityRole()

	rule, exists := l.rules[role]
	if !exists {
		rule = l.rules[model.UserRoleGuest]
	}

	if rule.MaxRequests == 0 {
		return errs.ErrRoleLimitExceeded
	}

	key := l.getCacheKey(user, role)

	currentVal, err := l.counter.Increment(ctx, key, rule.Window)
	if err != nil {
		return fmt.Errorf("rate limit check failed: %w", err)
	}

	if currentVal > int64(rule.MaxRequests) {
		return errs.ErrRoleLimitExceeded
	}

	return nil
}

func (l *RoleLimiter) Rollback(ctx context.Context, user model.User) error {
	role := user.HighestPriorityRole()
	rule, exists := l.rules[role]
	if !exists {
		rule = l.rules[model.UserRoleGuest]
	}

	if rule.MaxRequests == 0 {
		return nil
	}

	key := l.getCacheKey(user, role)
	if err := l.counter.Decrement(ctx, key); err != nil {
		return fmt.Errorf("rate limit rollback failed: %w", err)
	}

	return nil
}

func (l *RoleLimiter) getCacheKey(user model.User, r model.UserRole) string {
	return fmt.Sprintf("orders_rate_limit:%s:%s", user.UUID, r)
}
