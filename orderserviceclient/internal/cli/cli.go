package cli

import "github.com/spf13/cobra"

type Cli struct {
	rootCmd *cobra.Command
}

func (c *Cli) Execute() error {
	return c.rootCmd.Execute()
}

func New() *Cli {
	c := &Cli{}

	var grpcAddr string
	var userUuid string

	rootCmd := &cobra.Command{
		Use:   "ordercli",
		Short: "Client for working with order service",
	}
	rootCmd.PersistentFlags().StringVarP(&grpcAddr, "addr", "a", "", "order-service gRPC endpoint (required)")
	rootCmd.PersistentFlags().StringVarP(&userUuid, "uid", "u", "", "user uuid (required)")

	rootCmd.AddCommand(CreateCmd())

	c.rootCmd = rootCmd
	return c
}
