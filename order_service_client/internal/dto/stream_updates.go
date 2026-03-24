package dto

import "fmt"

type StreamOrderUpdateDto struct {
	OrderUUID string
	UserUUID  string
}

func (d *StreamOrderUpdateDto) Validate() error {
	if d.UserUUID == "" {
		return fmt.Errorf("empty user uuid")
	}

	if d.OrderUUID == "" {
		return fmt.Errorf("empty order uuid")
	}

	return nil
}
