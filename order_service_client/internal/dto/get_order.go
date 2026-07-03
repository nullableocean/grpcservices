package dto

import "fmt"

type GetOrderParams struct {
	OrderUUID string
}

func (d *GetOrderParams) Validate() error {
	if d.OrderUUID == "" {
		return fmt.Errorf("empty order uuid")
	}

	return nil
}
