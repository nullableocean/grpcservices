package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateCmd() *cobra.Command {
	var (
		marketUUID string
		orderType  string
		price      string
		quantity   int64
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create order and streaming updates",
		Run: func(cmd *cobra.Command, args []string) {
			addr, _ := cmd.Flags().GetString("addr")
			userUUID, _ := cmd.Flags().GetString("uid")

			inters := grpc.WithChainUnaryInterceptor(intercepter.UnaryClientXReqId())
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), inters)
			if err != nil {
				log.Fatalf("grpc connection error: %v", err)
			}
			defer conn.Close()

			client := orderv1.NewOrderClient(conn)

			var protoType typesv1.OrderType
			switch orderType {
			case "buy":
				protoType = typesv1.OrderType_ORDER_TYPE_BUY
			case "sell":
				protoType = typesv1.OrderType_ORDER_TYPE_SELL
			default:
				log.Fatalf("invalid order type: %s (wait buy/sell)", orderType)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			priceDec, err := decimal.NewFromString(price)
			if err != nil {
				log.Fatalf("invalid price: %v", err)
			}

			resp, err := client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
				UserUuid:  userUUID,
				MarketId:  marketUUID,
				OrderType: protoType,
				Price:     mapDecimalToProtoMoney(priceDec),
				Quantity:  quantity,
			})
			if err != nil {
				log.Fatalf("failed create order: %v", err)
			}

			fmt.Printf("%s %s\n\n", resp.OrderUuid, order.OrderStatus(resp.Status).String())
			fmt.Println("connecting to streaming updates...")

			streamCtx, streamCancel := context.WithCancel(context.Background())
			defer streamCancel()

			stream, err := client.StreamOrderUpdates(streamCtx, &orderv1.GetStatusRequest{
				OrderUuid: resp.OrderUuid,
				UserUuid:  userUUID,
			})
			if err != nil {
				log.Fatalf("failed stream connect: %v", err)
			}

			fmt.Printf("connected. updates:\n\n")
			go func() {
				for {
					resp, err := stream.Recv()
					if err != nil {
						if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
							fmt.Println("\nclosed")
							os.Exit(0)
						}

						if errors.Is(err, io.EOF) {
							fmt.Println("\nclosed by server")
							os.Exit(0)
						}

						fmt.Println("\nclosed with error", err)
						os.Exit(1)
					}

					fmt.Printf("%s\n", order.OrderStatus(resp.Status).String())
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			streamCancel()
		},
	}

	cmd.Flags().StringVarP(&marketUUID, "market", "m", "", "market UUID (required)")
	cmd.Flags().StringVarP(&orderType, "type", "t", "", "order type: buy/sell (required)")
	cmd.Flags().StringVarP(&price, "price", "p", "0", "price float (required)")
	cmd.Flags().Int64VarP(&quantity, "quantity", "q", 0, "position quantity (required)")

	cmd.MarkFlagRequired("market")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("price")
	cmd.MarkFlagRequired("quantity")

	return cmd
}

func mapDecimalToProtoMoney(dec decimal.Decimal) *typesv1.Money {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &typesv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}
