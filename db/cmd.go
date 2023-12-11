package db

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/log"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate data to database",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.New()
		if err != nil {
			panic(err)
		}

		db, err := New(logger)
		if err != nil {
			panic(err)
		}

		if err := migrateData(logger, db); err != nil {
			panic(err)
		}
		os.Exit(0)
	},
}

func init() {
	fmt.Println("init db")
	configFile := cmd.RootCmd.PersistentFlags().StringP("config", "c", "config", "config file name")
	viper.BindPFlag("config", cmd.RootCmd.PersistentFlags().Lookup("config"))
	viper.SetConfigName(*configFile)
	viper.AddConfigPath(".")
	cmd.RootCmd.AddCommand(MigrateCmd)
}

func migrateData(logger *zap.Logger, db *TiDb) error {
	conn := db.GetConnRoundRobin()
	driver, err := mysql.WithInstance(conn, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %w", err)
	}
	logger.Info("Running migration...")
	if err := m.Up(); err != nil {
		return fmt.Errorf("error running migration: %w", err)
	}
	logger.Info("Running migration successfully")
	return nil
}
