package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/client"
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

	cmd.Flags().StringVarP(&c.args.CreateArgs.MarketUUID, "market", "m", "", "market UUID (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.OrderSide, "side", "s", "", "order type: [buy|sell] (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.OrderType, "type", "t", "", "order type: [limit|market|stop|profit] (required)")
	cmd.Flags().StringVarP(&c.args.CreateArgs.Price, "price", "p", "0", "price float (required)")
	cmd.Flags().Int64VarP(&c.args.CreateArgs.Quantity, "quantity", "q", 0, "position quantity(required)")
	cmd.Flags().BoolVarP(&c.args.CreateArgs.WithStream, "stream", "", false, "with updates stream for created order")

	cmd.MarkFlagRequired("market")
	cmd.MarkFlagRequired("side")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("args.CreateArgs.Price")
	cmd.MarkFlagRequired("args.CreateArgs.Quantity")

	c.setCreateOrderRunFunc(cmd)

	return cmd
}

func (c *Cli) setCreateOrderRunFunc(cmd *cobra.Command) {
	cmd.Run = func(cmd *cobra.Command, args []string) {
		client := c.getOrderClient()

		token := c.args.User.Jwt
		userUUID, err := c.parseUserFromToken(token)
		if err != nil {
			log.Fatalln("failed parse token:", err)
		}

		orderUUID := c.createOrder(client, token, userUUID, c.args.CreateArgs)

		fmt.Println("OK", orderUUID)

		if c.args.CreateArgs.WithStream {
			c.streamUpdates(client, token, orderUUID, userUUID)
		}
	}
}

func (c *Cli) parseUserFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}

	user, ex := claims["sub"].(string)
	if !ex {
		return "", errors.New("sub not found in token")
	}

	return user, nil
}

func (c *Cli) createOrder(orderClient *client.Client, token, userUUID string, args CreateArgs) string {
	priceDec, err := decimal.NewFromString(args.Price)
	if err != nil {
		log.Fatalf("invalid args.CreateArgs.Price: %v", err)
	}

	createDto := &dto.CreateOrderParameters{
		UserUUID:   userUUID,
		MarketUUID: args.MarketUUID,
		Type:       model.OrderType(args.OrderType),
		Side:       model.OrderSide(args.OrderSide),
		Price:      priceDec,
		Quantity:   decimal.NewFromInt(args.Quantity),
	}

	resp, err := orderClient.CreateOrder(context.Background(), token, createDto)
	if err != nil {
		log.Fatalln("failed create order", err)
	}

	orderUUID := resp.NewOrderUuid

	return orderUUID
}
