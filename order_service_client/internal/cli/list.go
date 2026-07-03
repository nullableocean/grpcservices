package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/client"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"github.com/spf13/cobra"
)

func (c *Cli) ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "user list orders",
	}

	cmd.Flags().StringVarP(&c.args.ListArgs.NextPageToken, "pageToken", "t", "", "Next Page Token")
	cmd.Flags().IntVarP(&c.args.ListArgs.PageSize, "listsize", "l", 10, "Page size (default 10)")

	cmd.Flags().StringSliceVar(&c.args.ListArgs.Filters.MarketUuids, "markets", []string{}, "filter orders by markets")
	cmd.Flags().StringSliceVar(&c.args.ListArgs.Filters.Statuses, "statuses", []string{}, "filter orders by statuses (--statuses=created,pending,completed,canceled)")
	cmd.Flags().StringVar(&c.args.ListArgs.Filters.Type, "ftype", "", "filter orders by type")
	cmd.Flags().TimeVar(&c.args.ListArgs.Filters.CreatedFrom, "from", time.Time{}, []string{time.DateTime, time.DateOnly}, "filter by created date from")
	cmd.Flags().TimeVar(&c.args.ListArgs.Filters.CreatedTo, "to", time.Time{}, []string{time.DateTime, time.DateOnly}, "filter by created date to")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		token := c.args.User.Jwt

		c.executeListOrders(token, c.args.ListArgs)
	}

	return cmd
}

func (c *Cli) executeListOrders(jwt string, args ListOrdersArgs) {
	listDto := &dto.ListOrdersParams{
		Filters:       c.getFiltersDto(args.Filters),
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

func (c *Cli) getFiltersDto(cmdFilters ListFiltersArgs) dto.Filters {
	filters := dto.Filters{}

	for _, s := range cmdFilters.Statuses {
		filters.Statuses = append(filters.Statuses, model.OrderStatus(s))
	}
	filters.MarketUuids = cmdFilters.MarketUuids

	if cmdFilters.Type != "" {
		t := model.OrderType(cmdFilters.Type)
		filters.Type = &t
	}

	if !cmdFilters.CreatedFrom.IsZero() {
		filters.CreatedFrom = &cmdFilters.CreatedFrom
	}

	if !cmdFilters.CreatedTo.IsZero() {
		filters.CreatedTo = &cmdFilters.CreatedTo
	}

	return filters
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
