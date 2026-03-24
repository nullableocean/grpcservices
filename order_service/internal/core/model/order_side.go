package model

import (
	"encoding/json"
	"fmt"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}

func (s OrderSide) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *OrderSide) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	side := OrderSide(str)
	if !side.IsValid() {
		return fmt.Errorf("invalid order side: %s", str)
	}

	*s = side
	return nil
}
