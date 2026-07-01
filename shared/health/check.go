package health

import "context"

// HealthCheck интерфейс для реализации проверки здоровья компонентов
type HealthCheck interface {
	Name() string
	Health(context.Context) error
}

type healthAdapter struct {
	name string
	fn   func(context.Context) error
}

func (ha *healthAdapter) Name() string {
	return ha.name
}

func (ha *healthAdapter) Health(ctx context.Context) error {
	return ha.fn(ctx)
}

// NewHealthCheck создает адаптер под интерфейс HealthCheck
//
// healthFn вызывается для проверки здоровья компонента
func NewHealthCheck(componentName string, healthFn func(context.Context) error) HealthCheck {
	return &healthAdapter{
		name: componentName,
		fn:   healthFn,
	}
}
