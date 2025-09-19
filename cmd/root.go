package cmd

import (
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gonstrukt",
		Short: "A CLI tool to spawn Go services with different configurations",
		Long: `gonstrukt is a CLI tool that generates Go microservices with different types
(gateway, auth service) and configurations including databases, caching,
observability, and rate limiting.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.AddCommand(CreateCmd())
	rootCmd.AddCommand(CompletionCmd())

	return rootCmd
}
