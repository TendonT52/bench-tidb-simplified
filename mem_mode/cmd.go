package memmode

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/load"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/log"
)

var memModeCmd = &cobra.Command{
	Use:   "mem",
	Short: "In memory mode",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.New()
		if err != nil {
			panic(err)
		}

		db, err := db.New(logger)
		if err != nil {
			panic(err)
		}

		loader, err := load.New(logger)
		if err != nil {
			panic(err)
		}

		memMode := New(db, logger)
		err = memMode.SaveAllAccountsToMem()
		if err != nil {
			panic(err)
		}

		memMode.CreateWorker()

		err = loader.StartLoadTest(logger, memMode, db)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	memModeCmd.Flags().DurationP("timeLimit", "t", 100*time.Millisecond, "limit time to wait for insert")
	viper.BindPFlag("timeLimit", memModeCmd.Flags().Lookup("timeLimit"))
	memModeCmd.Flags().Int32P("sizeLimit", "s", 100, "limit size to wait for insert")
	viper.BindPFlag("sizeLimit", memModeCmd.Flags().Lookup("sizeLimit"))
	memModeCmd.Flags().DurationP("updateDuration", "u", 10*time.Second, "update duration")
	viper.BindPFlag("updateDuration", memModeCmd.Flags().Lookup("updateDuration"))
	memModeCmd.Flags().Int32P("numberQueue", "q", 1, "number of queue")
	viper.BindPFlag("numberQueue", memModeCmd.Flags().Lookup("numberQueue"))

	cmd.RootCmd.AddCommand(memModeCmd)
}
