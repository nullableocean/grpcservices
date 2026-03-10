package cli

import (
	"github.com/spf13/cobra"
)

type args struct {
	grpcAddr   string
	userUuid   string
	marketUUID string
	orderType  string
	price      string
	quantity   int64
}

type Cli struct {
	rootCmd *cobra.Command

	args
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
	rootCmd.PersistentFlags().StringVarP(&c.args.grpcAddr, "addr", "a", "", "order-service gRPC endpoint (required)")
	rootCmd.PersistentFlags().StringVarP(&c.args.userUuid, "uid", "u", "", "user uuid (required)")

	rootCmd.MarkFlagRequired("addr")
	rootCmd.MarkFlagRequired("uid")

	rootCmd.AddCommand(c.CreateCmd())

	c.rootCmd = rootCmd
	return c
}
