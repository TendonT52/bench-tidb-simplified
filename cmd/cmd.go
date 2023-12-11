package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "tidb-bench",
	Short: "TiDB benchmark tool",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func ReadCmdArg() error {
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %w", err))
	}

	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}
	return nil
}
