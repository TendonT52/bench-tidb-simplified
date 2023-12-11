package load

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
)

type Transfer interface {
	Transfer(from string, to string, amount int) error
}

type Avg struct {
	NumUser int
	Sum     int
	max     int
	min     int
}

func (a Avg) GetAvg() int {
	return int(a.Sum) / int(a.NumUser)
}

func (l *Loader) StartLoadTest(logger *zap.Logger, transfer Transfer, db *db.TiDb) error {
	accountIds := make([]string, l.LoadTestConfig.NumberOfUsersSet)
	logger.Info("Getting account ids...")
	err := l.getAccountId(db, accountIds)
	if err != nil {
		logger.Sugar().Errorf("error getting account id: %v", err)
		return err
	}
	// go func() {
	// 	err := l.getAccountId(db, accountIds)
	// 	if err != nil {
	// 		logger.Sugar().Errorf("error getting account id: %v", err)
	// 		return
	// 	}
	// }()

	// filledAccountIds := make(chan bool)
	// go func() {
	// 	for accountIds[len(accountIds)-1] == "" {
	// 		time.Sleep(100 * time.Millisecond)
	// 	}
	// 	close(filledAccountIds)
	// }()
	// <-filledAccountIds
	logger.Info("Get account ids done")

	logger.Info("Load test starting...")
	ctx, cancel := context.WithTimeout(context.Background(), l.LoadTestConfig.Duration)
	defer cancel()

	var stat []map[int]Avg

	for i := 0; i < l.LoadTestConfig.NumVirtualUsers; i++ {
		go func(i int) {
			m := make(map[int]Avg)
			stat = append(stat, m)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					err := l.GetStat(ctx, logger, transfer, accountIds, m)
					if err != nil {
						logger.Sugar().Errorf("error getting stat: %v", err)
					}
				}
			}
		}(i)
	}

	<-ctx.Done()

	result := make(map[int]Avg)

	for _, m := range stat {
		logger.Sugar().Debugf("Stat: %v", m)
		for k, v := range m {
			if v.NumUser == 0 {
				continue
			}
			if _, ok := result[k]; !ok {
				result[k] = Avg{
					max: v.max,
					min: v.min,
				}
			} else {
				avg := result[k]
				avg.NumUser += v.NumUser
				avg.Sum += v.Sum
				avg.max = max(avg.max, v.max)
				avg.min = min(avg.min, v.min)
				result[k] = avg
			}
		}
	}

	sumTransfer := 0
	for k, v := range result {
		sumTransfer += v.NumUser
		logger.Sugar().Infof("Result: %v collision, min: %v ms, avg: %v ms, max: %v ms", k, v.min, v.GetAvg(), v.max)
	}
	logger.Sugar().Infof("Total transfer: %v", sumTransfer)
	logger.Sugar().Infof("Transfer per second: %v", sumTransfer/int(l.LoadTestConfig.Duration.Seconds()))

	return nil
}

func (l *Loader) GetStat(ctx context.Context, logger *zap.Logger, transfer Transfer, accountIds []string, m map[int]Avg) error {
	from := accountIds[rand.Intn(len(accountIds))]
	to := accountIds[rand.Intn(len(accountIds))]
	collision := l.RandomCollision()
	collSide := rand.Intn(2)
	amount := rand.Intn(100)
	if amount == 0 {
		amount = 1
	}
	for i := 0; i < collision; i++ {
		var duration time.Duration
		if collSide == 0 {
			from = accountIds[rand.Intn(len(accountIds))]
			if from == to {
				continue
			}
			start := time.Now()
			err := transfer.Transfer(from, to, amount)
			duration = time.Since(start)
			if err != nil {
				return fmt.Errorf("error transferring: %w", err)
			}
		} else {
			to = accountIds[rand.Intn(len(accountIds))]
			if from == to {
				continue
			}
			start := time.Now()
			err := transfer.Transfer(to, from, amount)
			duration = time.Since(start)
			if err != nil {
				return fmt.Errorf("error transferring: %w", err)
			}
		}
		if v, ok := m[collision]; ok {
			v.NumUser++
			v.Sum += int(duration.Milliseconds())
			v.max = max(v.max, int(duration.Milliseconds()))
			v.min = min(v.min, int(duration.Milliseconds()))
			m[collision] = v
		} else {
			m[collision] = Avg{
				NumUser: 1,
				Sum:     int(duration.Milliseconds()),
				max:     int(duration.Milliseconds()),
				min:     int(duration.Milliseconds()),
			}
		}
	}
	return nil
}

func (l *Loader) getAccountId(db *db.TiDb, accountIds []string) error {
	rows, err := db.GetConnRoundRobin().Query("SELECT id, balance FROM account")
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var id string
		var balance int
		err := rows.Scan(&id, &balance)
		if err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		if balance == 0 {
			continue
		}
		accountIds[i] = id
		if i == len(accountIds)-1 {
			break
		}
		i = (i + 1) % len(accountIds)
	}

	return nil
}

func (l *Loader) RandomCollision() int {
	randomNum := rand.Intn(1000000)
	sum := 0
	for _, collision := range l.CollisionRates {
		sum += collision.numUser
		if randomNum < sum {
			return collision.collision
		}
	}
	return 1
}
