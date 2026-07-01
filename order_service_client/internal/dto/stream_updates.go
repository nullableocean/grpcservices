package dto

import "fmt"

type StreamUpdatesParams struct {
	OrderUUID string
	UserUUID  string
}

func (d *StreamUpdatesParams) Validate() error {
	if d.UserUUID == "" {
		return fmt.Errorf("empty user uuid")
	}

	if d.OrderUUID == "" {
		return fmt.Errorf("empty order uuid")
	}

	return nil
}
