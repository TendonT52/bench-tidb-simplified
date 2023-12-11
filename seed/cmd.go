package seed

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/load"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/log"
)

var numAccount *int
var balance *int

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed data to database",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := log.New()
		if err != nil {
			return fmt.Errorf("error creating logger: %w", err)
		}

		db, err := db.New(logger)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		loader, err := load.New(logger)
		if err != nil {
			return fmt.Errorf("error creating loader: %w", err)
		}

		SeedAccount(logger, loader, db)

		return nil
	},
}

func init() {
	numAccount = seedCmd.Flags().IntP("accountNumber", "n", 10000, "number of account")
	balance = seedCmd.Flags().IntP("balance", "b", 1000000, "balance of each account")
	cmd.RootCmd.AddCommand(seedCmd)
}

func readSeedConfig(logger *zap.Logger) error {
	logger.Info("Reading seed config")
	logger.Sugar().Infof("Number of account: %v", *numAccount)
	if *numAccount == 0 {
		return fmt.Errorf("number of account must be greater than 0")
	}
	logger.Info("Read seed config successfully")
	return nil
}
