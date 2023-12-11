package seed

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/load"
)

func SeedAccount(logger *zap.Logger, loader *load.Loader, db *db.TiDb) error {

	readSeedConfig(logger)

	logger.Info("Seeding account...")
	logger.Sugar().Infof("Number of account: %v", *numAccount)
	logger.Sugar().Infof("Balance of each account: %v", *balance)
	logger.Sugar().Infof("Number of virtual users: %d", loader.LoadTestConfig.NumVirtualUsers)
	logger.Sugar().Infof("Number of account per virtual user: %d", *numAccount/loader.LoadTestConfig.NumVirtualUsers)
	logger.Sugar().Infof("Time out %v", loader.LoadTestConfig.Duration)

	var wg sync.WaitGroup
	wg.Add(loader.LoadTestConfig.NumVirtualUsers)

	ctx, cancel := context.WithTimeout(context.Background(), loader.LoadTestConfig.Duration)
	defer cancel()

	for i := 0; i < loader.LoadTestConfig.NumVirtualUsers; i++ {
		go func(wg *sync.WaitGroup, ctx context.Context) {
			defer wg.Done()
			err := InsertAccount(ctx, *numAccount/loader.LoadTestConfig.NumVirtualUsers, db)
			if err != nil {
				cancel()
				logger.Sugar().Errorf("Error inserting account: %v", err)
			}
		}(&wg, ctx)
	}

	wg.Wait()

	return nil
}

func InsertAccount(ctx context.Context, nAccount int, db *db.TiDb) error {
	const batchSize = 500

	for i := 0; i < nAccount; i += batchSize {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context timeout")
		default:
			db := db.GetConnRoundRobin()

			valueStrings := make([]string, 0, batchSize)
			valueArgs := make([]interface{}, 0, batchSize)

			for j := 0; j < batchSize && i+j < nAccount; j++ {
				valueStrings = append(valueStrings, "(?, ?)")
				valueArgs = append(valueArgs, ulid.Make().String(), *balance)
			}

			insertQuery := fmt.Sprintf("INSERT INTO account (id, balance) VALUES %s", strings.Join(valueStrings, ","))
			_, err := db.Exec(insertQuery, valueArgs...)
			if err != nil {
				return fmt.Errorf("error inserting account: %w", err)
			}
		}
	}
	return nil
}
