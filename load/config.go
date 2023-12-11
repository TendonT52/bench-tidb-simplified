package load

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"
)

func init() {
	cmd.RootCmd.PersistentFlags().StringP("num-virtual-user", "v", "100", "number of virtual users")
	viper.BindPFlag("num-virtual-user", cmd.RootCmd.PersistentFlags().Lookup("num-virtual-user"))
	cmd.RootCmd.PersistentFlags().StringP("number-of-users-set", "o", "100000", "number of users set")
	viper.BindPFlag("number-of-users-set", cmd.RootCmd.PersistentFlags().Lookup("number-of-users-set"))
	cmd.RootCmd.PersistentFlags().StringP("duration", "d", "1m", "duration of the test")
	viper.BindPFlag("duration", cmd.RootCmd.PersistentFlags().Lookup("duration"))
	cmd.RootCmd.PersistentFlags().StringP("collision-rate", "r", "1000000,1", "collision rate of the test")
	viper.BindPFlag("collision-rate", cmd.RootCmd.PersistentFlags().Lookup("collision-rate"))
}

type LoadConfig struct {
	NumVirtualUsers  int           `validate:"required" mapstructure:"num-virtual-user"`
	NumberOfUsersSet int           `validate:"required" mapstructure:"number-of-users-set"`
	Duration         time.Duration `validate:"required" mapstructure:"duration"`
	CollisionRate    string        `validate:"required" mapstructure:"collision-rate"`
}

type Loader struct {
	LoadTestConfig LoadConfig
	CollisionRates []collisions
}

var loader *Loader

func New(logger *zap.Logger) (*Loader, error) {
	if loader != nil {
		return loader, nil
	}

	loader = &Loader{}

	logger.Info("Load test config starting...")
	err := viper.Unmarshal(&loader.LoadTestConfig)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling load test config: %w", err)
	}
	logger.Debug(fmt.Sprintf("Load test config: %+v", loader.LoadTestConfig))
	validator := validator.New()
	err = validator.Struct(loader.LoadTestConfig)
	if err != nil {
		return nil, fmt.Errorf("error validating load test config: %w", err)
	}

	loader.CollisionRates, err = parsePairs(loader.LoadTestConfig.CollisionRate)
	if err != nil {
		return nil, fmt.Errorf("error parsing collision rate: %w", err)
	}

	logger.Info(fmt.Sprintf("Parsed collision rate: %v", loader.CollisionRates))
	return loader, nil
}

type collisions struct {
	numUser   int
	collision int
}

func parsePairs(pairInput string) ([]collisions, error) {
	pairs := strings.Split(pairInput, " ")
	result := make([]collisions, len(pairs))
	for i, pair := range pairs {
		nums := strings.Split(pair, ",")
		if len(nums) != 2 {
			return nil, fmt.Errorf("invalid pair: %s", pair)
		}
		nUser, err := strconv.Atoi(nums[0])
		if err != nil {
			return nil, fmt.Errorf("invalid collision: %s", pair)
		}
		nColl, err := strconv.Atoi(nums[1])
		if err != nil {
			return nil, fmt.Errorf("invalid collision: %s", pair)
		}
		result[i] = collisions{
			numUser:   nUser,
			collision: nColl,
		}
	}

	sum := 0
	for _, c := range result {
		if c.collision < 0 {
			return nil, fmt.Errorf("invalid collision: %d", c.collision)
		}
		if c.numUser < 0 {
			return nil, fmt.Errorf("invalid number of user collision: %d", c.numUser)
		}
		sum += c.numUser
	}

	if sum > 1000000 {
		return nil, fmt.Errorf("collision rate cannot be greater than 1000000(1M)")
	}

	if sum < 1000000 {
		result = append(result, collisions{
			numUser:   1000000 - sum,
			collision: 1,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].numUser < result[j].numUser
	})

	return result, nil
}
