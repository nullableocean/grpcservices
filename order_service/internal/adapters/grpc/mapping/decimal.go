package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/shopspring/decimal"
)

func MapProtoMoneyToDecimal(pbmoney *modelsv1.Money) decimal.Decimal {
	units := decimal.NewFromInt(pbmoney.Units)
	nanos := decimal.NewFromInt32(pbmoney.Nanos).Div(decimal.NewFromInt(1e9))

	return units.Add(nanos)
}

func MapProtoDecimalToDecimal(pbdecimal *modelsv1.Decimal) decimal.Decimal {
	units := decimal.NewFromInt(pbdecimal.Units)
	nanos := decimal.NewFromInt32(pbdecimal.Nanos).Div(decimal.NewFromInt(1e9))

	return units.Add(nanos)
}

func MapDecimalToProtoMoney(dec decimal.Decimal) *modelsv1.Money {
	units := dec.IntPart()

	fractional := dec.Sub(decimal.NewFromInt(units))
	nanos := fractional.Mul(decimal.NewFromInt(1e9)).IntPart()

	return &modelsv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}
