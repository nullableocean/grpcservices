package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

func (c *Cli) CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create order and streaming updates",
	}

	cmd.Flags().StringVarP(&c.args.CreateArgs.IdempotencyKey, "idemkey", "k", "", "idempotency key (UUID)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.MarketUUID, "market", "m", "", "market UUID (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.OrderSide, "side", "s", "", "order type: [buy|sell] (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.OrderType, "type", "t", "", "order type: [limit|market|stop|profit] (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.Price, "price", "p", "0", "price float (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.Quantity, "quantity", "q", "0.0", "position quantity float(required)")
	cmd.Flags().BoolVar(&c.args.CreateArgs.WithStream, "stream", false, "with updates stream for created order")

	cmd.MarkFlagRequired("market")
	cmd.MarkFlagRequired("side")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("price")
	cmd.MarkFlagRequired("quantity")

	c.setCreateOrderRunFunc(cmd)

	return cmd
}

func (c *Cli) setCreateOrderRunFunc(cmd *cobra.Command) {
	cmd.Run = func(cmd *cobra.Command, args []string) {

		token := c.args.User.Jwt
		userUUID := c.args.User.UUID

		orderUUID := c.createOrder(token, userUUID, c.args.CreateArgs)

		fmt.Println("OK", orderUUID)

		if c.args.CreateArgs.WithStream {
			c.streamUpdates(token, orderUUID, userUUID)
		}
	}
}

func (c *Cli) createOrder(token, userUUID string, args CreateArgs) string {
	priceDec, err := decimal.NewFromString(args.Price)
	if err != nil {
		log.Fatalf("invalid price: %v", err)
	}

	quantityDec, err := decimal.NewFromString(args.Quantity)
	if err != nil {
		log.Fatalf("invalid quantity %v", err)
	}

	createDto := &dto.CreateOrderParams{
		UserUUID:   userUUID,
		MarketUUID: args.MarketUUID,
		Type:       model.OrderType(args.OrderType),
		Side:       model.OrderSide(args.OrderSide),
		Price:      priceDec,
		Quantity:   quantityDec,
	}

	resp, err := c.getOrderClient().CreateOrder(context.Background(), token, createDto)
	if err != nil {
		log.Fatalln("failed create order", err)
	}

	orderUUID := resp.NewOrderUuid
	return orderUUID
}
