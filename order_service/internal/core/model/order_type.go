package model

import (
	"encoding/json"
	"fmt"
)

type OrderType string

const (
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeStopLoss   OrderType = "STOP_LOSS"
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT"
)

func (t OrderType) IsValid() bool {
	switch t {
	case OrderTypeLimit, OrderTypeMarket, OrderTypeStopLoss, OrderTypeTakeProfit:
		return true
	}
	return false
}

func (t OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *OrderType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	orderType := OrderType(str)
	if !orderType.IsValid() {
		return fmt.Errorf("invalid order type: %s", str)
	}

	*t = orderType
	return nil
}
