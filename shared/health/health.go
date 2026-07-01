package health

import (
	"context"
	"sync"
)

const (
	HEALTH    Status = "health"
	UNHEALTHY Status = "unhealthy"
)

type Status string

type ComponentResult struct {
	ComponentName string `json:"name"`
	Status        Status `json:"status"`
}

type HealthResult struct {
	// Итоговый статус
	Status Status `json:"status"`
	// Результаты отдельных HealthCheck
	Components []*ComponentResult `json:"components"`
}

type HealthCheker struct {
	checks []HealthCheck
}

func NewHealthCheker(checks ...HealthCheck) *HealthCheker {
	return &HealthCheker{
		checks: checks,
	}
}

func (c *HealthCheker) Add(check HealthCheck) {
	c.checks = append(c.checks, check)
}

// Health обходит добавленные HealthCheck
func (c *HealthCheker) Health(ctx context.Context) *HealthResult {
	result := &HealthResult{
		Status:     HEALTH,
		Components: make([]*ComponentResult, len(c.checks)),
	}

	wg := sync.WaitGroup{}
	for i, healthcheck := range c.checks {
		wg.Add(1)
		go func() {
			defer wg.Done()

			componentResult := &ComponentResult{
				ComponentName: healthcheck.Name(),
			}

			err := healthcheck.Health(ctx)
			if err != nil {
				componentResult.Status = UNHEALTHY
			} else {
				componentResult.Status = HEALTH
			}

			result.Components[i] = componentResult
		}()
	}
	wg.Wait()

	for _, c := range result.Components {
		if c.Status == UNHEALTHY {
			result.Status = UNHEALTHY
			break
		}
	}

	return result
}
