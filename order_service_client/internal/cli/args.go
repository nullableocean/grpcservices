package cli

import "time"

type UserArgs struct {
	Jwt  string
	UUID string
}

type CreateArgs struct {
	MarketUUID     string
	OrderType      string
	OrderSide      string
	Price          string
	Quantity       string
	WithStream     bool
	IdempotencyKey string
}

type ListOrdersArgs struct {
	PageSize      int
	NextPageToken string
	Filters       ListFiltersArgs
}

type ListFiltersArgs struct {
	Statuses    []string
	MarketUuids []string
	Type        string
	CreatedFrom time.Time
	CreatedTo   time.Time
}

type StreamArgs struct {
	OrderUUID string
}

type GetOrderArgs struct {
	OrderUUID string
}

type Args struct {
	GrpcAddr string
	User     UserArgs

	StreamArgs StreamArgs
	CreateArgs CreateArgs
	ListArgs   ListOrdersArgs
	GetOrder   GetOrderArgs
}
