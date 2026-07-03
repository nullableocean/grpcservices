package dto

import (
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
)

type Filters struct {
	Statuses    []model.OrderStatus
	MarketUuids []string
	Type        *model.OrderType
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

func (f Filters) Validate() error {
	if len(f.Statuses) > 0 {
		for _, s := range f.Statuses {
			if !s.IsValid() {
				return fmt.Errorf("undefined status: %s", s)
			}
		}
	}

	if f.Type != nil && !f.Type.IsValid() {
		return fmt.Errorf("undefined order type: %s", *f.Type)
	}

	return nil
}

type ListOrdersParams struct {
	PageSize      int
	NextPageToken string
	Filters       Filters
}

func (p *ListOrdersParams) Validate() error {
	if p.PageSize <= 0 {
		return fmt.Errorf("page size negative or zero")
	}

	return nil
}
