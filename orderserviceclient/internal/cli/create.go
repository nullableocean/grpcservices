package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/client"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

func (c *Cli) CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create order and streaming updates",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.NewClient(c.grpcAddr)
			if err != nil {
				log.Fatalln("create client error", err)
			}

			var ot order.OrderType
			switch c.orderType {
			case "buy":
				ot = order.ORDER_TYPE_BUY
			case "sell":
				ot = order.ORDER_TYPE_SELL
			default:
				log.Fatalf("invalid order type: %s (wait buy/sell)\n", c.orderType)
			}

			priceDec, err := decimal.NewFromString(c.price)
			if err != nil {
				log.Fatalf("invalid price: %v", err)
			}

			userUuid, _ := cmd.Flags().GetString("uid")
			createDto := &dto.CreateOrderDto{
				OrderType:  ot,
				UserUuid:   userUuid,
				MarketUuid: c.marketUUID,
				Price:      priceDec,
				Quantity:   decimal.NewFromInt(c.quantity),
			}
			resp, err := client.CreateOrder(context.Background(), createDto)
			if err != nil {
				log.Fatalln("failed create order", err)
			}
			orderUuid := resp.NewOrderUuid
			fmt.Printf("%s %s\n\n", orderUuid, resp.Status.String())

			fmt.Println("connecting to streaming updates...")
			streamCtx, streamCancel := context.WithCancel(context.Background())
			defer streamCancel()

			dataCh, err := client.StreamOrderUpdates(streamCtx, &dto.StreamOrderUpdateDto{
				OrderUuid: orderUuid,
				UserUuid:  userUuid,
			})
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Printf("connected. updates:\n\n")
			go func() {
				for data := range dataCh {
					if data.Err != nil {
						fmt.Println(data.Err.Error())
						continue
					}

					fmt.Println(data.NewStatus.String())
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			streamCancel()
		},
	}

	cmd.Flags().StringVarP(&c.args.marketUUID, "market", "m", "", "market UUID (required)")
	cmd.Flags().StringVarP(&c.args.orderType, "type", "t", "", "order type: buy/sell (required)")
	cmd.Flags().StringVarP(&c.args.price, "price", "p", "0", "price float (required)")
	cmd.Flags().Int64VarP(&c.args.quantity, "quantity", "q", 0, "position quantity (required)")

	cmd.MarkFlagRequired("market")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("price")
	cmd.MarkFlagRequired("quantity")

	return cmd
}
