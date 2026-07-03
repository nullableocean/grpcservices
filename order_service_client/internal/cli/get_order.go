package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/spf13/cobra"
)

func (c *Cli) GetOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get order by uuid",
	}

	cmd.Flags().StringVar(&c.args.GetOrder.OrderUUID, "order", "", "order uuid")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		token := c.args.User.Jwt
		c.executeGetOrder(token, c.args.GetOrder)
	}

	return cmd
}

func (c *Cli) executeGetOrder(jwt string, args GetOrderArgs) {
	getDto := &dto.GetOrderParams{
		OrderUUID: args.OrderUUID,
	}

	client := c.getOrderClient()
	resp, err := client.GetOrder(context.Background(), jwt, getDto)
	if err != nil {
		log.Fatalln("failed get order", err)
	}

	fmt.Println("ORDER")
	fmt.Println(resp.Order.String())
}
