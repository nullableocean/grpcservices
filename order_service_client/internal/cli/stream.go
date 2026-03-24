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
	"github.com/spf13/cobra"
)

func (c *Cli) StreamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "stream status updates for order",
	}

	cmd.Flags().StringVarP(&c.args.StreamArgs.OrderUUID, "order", "o", "", "order UUID (required)")
	cmd.MarkFlagRequired("order")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		token := c.args.User.Jwt
		userUUID, err := c.parseUserFromToken(token)
		if err != nil {
			log.Fatalln("failed parse token:", err)
		}

		c.streamUpdates(c.getOrderClient(), token, c.args.StreamArgs.OrderUUID, userUUID)
	}

	return cmd
}

func (c *Cli) streamUpdates(orderClient *client.Client, token, orderUUID, userUUID string) {
	streamCtx, streamCancel := context.WithCancel(context.Background())
	defer streamCancel()

	dataCh, err := orderClient.StreamOrderUpdates(streamCtx, token, &dto.StreamOrderUpdateDto{
		OrderUUID: orderUUID,
		UserUUID:  userUUID,
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

			fmt.Println(data.NewStatus)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	streamCancel()
}
