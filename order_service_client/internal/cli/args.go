package cli

type UserArgs struct {
	Jwt string
}

type CreateArgs struct {
	MarketUUID string
	OrderType  string
	OrderSide  string
	Price      string
	Quantity   int64
	WithStream bool
}

type StreamArgs struct {
	OrderUUID string
}

type Args struct {
	GrpcAddr string
	User     UserArgs

	StreamArgs StreamArgs
	CreateArgs CreateArgs
}
