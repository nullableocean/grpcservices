package client

import (
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/shopspring/decimal"
)

func mapDecimalToProtoMoney(dec decimal.Decimal) *typesv1.Money {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &typesv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}
