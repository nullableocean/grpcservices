package cli

import (
	"log"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/client"
	"github.com/spf13/cobra"
)

type Cli struct {
	rootCmd *cobra.Command
	args    Args
}

func (c *Cli) Execute() error {
	return c.rootCmd.Execute()
}

func New() *Cli {
	c := &Cli{}

	rootCmd := &cobra.Command{
		Use:   "ordercli",
		Short: "Client for working with order service",
	}
	rootCmd.PersistentFlags().StringVarP(&c.args.GrpcAddr, "addr", "a", "", "order-service gRPC endpoint (required)")
	rootCmd.PersistentFlags().StringVarP(&c.args.User.Jwt, "jwt", "j", "", "jwt token (required)")

	rootCmd.MarkFlagRequired("addr")
	rootCmd.MarkFlagRequired("jwt")

	rootCmd.AddCommand(c.CreateCmd())
	rootCmd.AddCommand(c.StreamCmd())

	c.rootCmd = rootCmd
	return c
}

func (c *Cli) getOrderClient() *client.Client {
	client, err := client.NewClient(c.args.GrpcAddr)
	if err != nil {
		log.Fatalln("create client error", err)
	}

	return client
}
