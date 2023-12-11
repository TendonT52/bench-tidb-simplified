package db

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type databaseConfig struct {
	Address      []string      `validate:"required" mapstructure:"addresses"`
	Username     string        `validate:"required" mapstructure:"username"`
	Password     string        `validate:"required" mapstructure:"password"`
	DatabaseName string        `validate:"required" mapstructure:"dbName"`
	MaxOpenConns int           `validate:"required" mapstructure:"maxOpenConns"`
	MaxIdleConns int           `validate:"required" mapstructure:"maxIdleConns"`
	ConnMaxIdle  time.Duration `validate:"required" mapstructure:"connMaxIdleTime"`
	ConnMaxLife  time.Duration `validate:"required" mapstructure:"connMaxLifeTime"`
	TLS          string        `mapstructure:"tls"`
}

func ReadDatabaseConfig(logger *zap.Logger) (databaseConfig, error) {
	var dbConfig databaseConfig

	viper.SetConfigName(viper.GetString("config"))

	logger.Sugar().Infof("Reading database config %s", viper.GetString("config"))
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	err = viper.UnmarshalKey("database", &dbConfig)
	if err != nil {
		return databaseConfig{}, fmt.Errorf("error unmarshalling database config: %w", err)
	}
	logger.Info("Read database config successfully")
	logger.Sugar().Infof("Database config: %+v", dbConfig)

	logger.Info("Validating database config")
	validate := validator.New()
	err = validate.Struct(dbConfig)
	if err != nil {
		return databaseConfig{}, fmt.Errorf("error validating database config: %w", err)
	}
	logger.Info("Validated database config successfully")

	return dbConfig, nil
}
