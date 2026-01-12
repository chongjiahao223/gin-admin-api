package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gin-api",
	Short: "gin-api 管理工具",
}

func init() {
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(cronCmd)
	taskCmd.AddCommand(runTaskCmd)
	rootCmd.AddCommand(taskCmd)
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
