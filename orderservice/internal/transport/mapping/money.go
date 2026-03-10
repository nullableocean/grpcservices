package mapping

import (
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/money"
	"github.com/shopspring/decimal"
)

// Map pb Money to domain money
func MapProtoMoneyToDomain(pbmoney *typesv1.Money) money.Money {
	return money.Money{
		Decimal: MapProtoMoneyToDecimal(pbmoney),
	}
}

// Map pb money to decimal struct
func MapProtoMoneyToDecimal(pbmoney *typesv1.Money) decimal.Decimal {
	units := decimal.NewFromInt(pbmoney.Units)
	nanos := decimal.NewFromInt(int64(pbmoney.Nanos))

	result := units.Add(nanos.Div(decimal.NewFromInt(1e9)))
	return result
}

func MapDomainMoneyToProto(m money.Money) *typesv1.Money {
	units := m.Decimal.IntPart()

	fractional := m.Decimal.Sub(decimal.NewFromInt(units))
	nanos := fractional.Mul(decimal.NewFromInt(1e9)).IntPart()

	return &typesv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}
