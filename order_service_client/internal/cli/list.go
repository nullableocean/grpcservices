package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/client"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/spf13/cobra"
)

func (c *Cli) ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "user list orders",
	}

	cmd.Flags().StringVarP(&c.args.ListArgs.NextPageToken, "pageToken", "t", "", "Next Page Token")
	cmd.Flags().IntVarP(&c.args.ListArgs.PageSize, "listsize", "l", 10, "Page size (default 10)")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		token := c.args.User.Jwt
		userUUID := c.args.User.UUID

		c.executeListOrders(token, userUUID, c.args.ListArgs)
	}

	return cmd
}

func (c *Cli) executeListOrders(jwt, userUUID string, args ListOrdersArgs) {
	listDto := &dto.ListOrdersParams{
		UserUUID:      userUUID,
		PageSize:      args.PageSize,
		NextPageToken: args.NextPageToken,
	}

	client := c.getOrderClient()
	listResp, err := client.ListOrders(context.Background(), jwt, listDto)
	if err != nil {
		log.Fatalln("failed get list orders", err)
	}

	c.printListResponse(listResp)
}

func (c *Cli) printListResponse(resp *client.ListResponse) {
	fmt.Println("Page count:", len(resp.Orders))
	if resp.NextPageToken != "" {
		fmt.Println("NextPage token:", len(resp.Orders))
	}

	fmt.Print("\n=== ORDERS ===\n\n")
	for _, o := range resp.Orders {
		fmt.Println(o.String())
	}
}
