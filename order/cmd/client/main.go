package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/pkg/intercepter"
	"github.com/nullableocean/grpcservices/pkg/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

/*
CLI Client

Консольный клиент для взаимодействия с сервисом Order
*/

type Command struct {
	grpcAddress string
	userId      int64
	marketId    int64
	price       float64
	quantity    int64
	orderType   order.OrderType
}

func (c *Command) validate() error {
	if c.grpcAddress == "" {
		return fmt.Errorf("invalid grpc endpoint")
	}
	if c.userId < 0 {
		return fmt.Errorf("invalid user id")
	}
	if c.marketId < 0 {
		return fmt.Errorf("invalid market id")
	}
	if c.price <= 0 {
		return fmt.Errorf("invalid price")
	}
	if c.quantity <= 0 {
		return fmt.Errorf("invalid quantity")
	}

	if order.MapOrderTypeToString(c.orderType) == "" {
		return fmt.Errorf("invalid order type")
	}

	return nil
}

func (c *Command) printOrderInfo() {
	format := "order info:\nuserid: %d market: %d price: %.4f quantity: %d type: %s\n\n"

	fmt.Printf(format,
		c.userId,
		c.marketId,
		c.price,
		c.quantity,
		order.MapOrderTypeToString(c.orderType),
	)
}

func main() {
	grpcAddress := flag.String("addr", "", "order service grpc endpoint")
	userId := flag.Int64("uid", -1, "user ID")
	marketId := flag.Int64("mid", -1, "market ID")
	price := flag.Float64("price", 0.0, "price for market position")
	quantity := flag.Int64("q", 0, "qunatity market position in order")
	orderType := flag.Int("type", 0, "order type. 1 - buy | 2 - sell")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -addr localhost:8085 -uid 1 -mid 2 -price 45 -q 1 -type 1 \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	cm := &Command{
		grpcAddress: *grpcAddress,
		userId:      *userId,
		marketId:    *marketId,
		price:       *price,
		quantity:    *quantity,
		orderType:   order.OrderType(*orderType),
	}

	if err := cm.validate(); err != nil {
		fmt.Printf("invalid args: %s\n\n", err.Error())
		flag.Usage()
		return
	}

	inters := grpc.WithChainUnaryInterceptor(
		intercepter.UnaryClientXReqId(),
	)
	conn, err := grpc.NewClient(cm.grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		inters,
	)
	if err != nil {
		fmt.Println("grpc connect error: ", err)
		return
	}

	client := orderpb.NewOrderClient(conn)

	cm.printOrderInfo()

	r, err := createOrder(client, cm)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			fmt.Println("create order request failed: ", err)
		}

		fmt.Printf("create order failed\n%s", s.Message())
		return
	}

	fmt.Printf("OK\nnew_order_id: %d status: %s\n\n", r.OrderId, order.MapOrderStatusToString(order.OrderStatus(r.Status)))
}

func createOrder(client orderpb.OrderClient, cm *Command) (*orderpb.CreateOrderResponse, error) {
	ctx := context.Background()

	request := &orderpb.CreateOrderRequest{
		UserId:    cm.userId,
		MarketId:  cm.marketId,
		OrderType: orderpb.OrderType(cm.orderType),
		Price:     float32(cm.price),
		Quantity:  cm.quantity,
	}

	return client.CreateOrder(ctx, request)
}
