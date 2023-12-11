package memmode

import (
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
)

type memmode struct {
	db             *db.TiDb
	logger         *zap.Logger
	currentSize    int32
	trigger        chan bool
	queuePool      *queuePool
	timeLimit      time.Duration
	sizeLimit      int
	updateDuration time.Duration
	lastInsert     time.Time
}

type LastTransferID struct {
	LastTransferID string
}

type Transaction struct {
	Ulid     string
	From     string
	To       string
	Amount   int32
	Response chan error
}

func New(db *db.TiDb, logger *zap.Logger) *memmode {
	numberQueue := viper.GetInt("numberQueue")
	sizeLimit := viper.GetInt("sizeLimit")
	timeLimit := viper.GetDuration("timeLimit")
	updateDuration := viper.GetDuration("updateDuration")

	logger.Sugar().Infof("sizeLimit: %d", sizeLimit)
	logger.Sugar().Infof("timeLimit: %v", timeLimit)

	return &memmode{
		db:             db,
		logger:         logger,
		queuePool:      newQueuePool(numberQueue, sizeLimit),
		timeLimit:      timeLimit,
		sizeLimit:      sizeLimit,
		updateDuration: updateDuration,
	}
}

func (n *memmode) SaveAllAccountsToMem() error {
	conn := n.db.GetConnRoundRobin()
	rows, err := conn.Query("SELECT id, balance FROM account")
	if err != nil {
		return err
	}
	defer rows.Close()

	n.logger.Sugar().Info("save all accounts to mem...")
	for rows.Next() {
		var id string
		var balance int
		err = rows.Scan(&id, &balance)
		if err != nil {
			return err
		}

		// only for limit account in memory for testing
		if balance == 0 {
			continue
		}
	}

	n.logger.Sugar().Info("save all accounts to mem done")

	return nil
}

func (n *memmode) Transfer(from string, to string, amount int) error {

	n.logger.Sugar().Debugf("transfer from %s to %s amount %d", from, to, amount)

	if from == to {
		return fmt.Errorf("from account id and to account id is the same")
	}
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	response := make(chan error, 1)

	transaction := &Transaction{
		Ulid:     ulid.Make().String(),
		From:     from,
		To:       to,
		Amount:   int32(amount),
		Response: response,
	}

	n.queuePool.getQueueRoundRobin() <- transaction

	err := <-response
	n.logger.Sugar().Debugf("transfer from %s to %s amount %d done", from, to, amount)
	return err
}

func (n *memmode) CreateWorker() {
	for i := 0; i < len(n.queuePool.pool); i++ {
		go n.BulkInsert(n.queuePool.pool[i])
	}
}

func (n *memmode) BulkInsert(queue queue) {
	transactions := make([]*Transaction, 0, n.sizeLimit)

loop:
	for i := 0; i < int(n.sizeLimit); i++ {
		select {
		case t := <-queue:
			transactions = append(transactions, t)
		case <-time.After(n.timeLimit):
			if len(transactions) > 0 {
				break loop
			}
			i--
		}
	}

	go n.BulkInsert(queue)

	n.logger.Sugar().Debugf("bulk insert %d transactions", len(transactions))

	valueStrings := make([]string, 0, len(transactions))
	valueArgs := make([]interface{}, 0, len(transactions))

	for _, t := range transactions {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, t.Ulid, t.From, t.To, t.Amount)
	}

	conn := n.db.GetConnRoundRobin()
	insertQuery := fmt.Sprintf("INSERT INTO transfer (ulid, from_account_id, to_account_id, amount) VALUES %s", strings.Join(valueStrings, ","))
	result, err := conn.Exec(insertQuery, valueArgs...)
	if err != nil {
		n.logger.Sugar().Errorf("error bulk insert: %v", err)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	if affected != int64(len(transactions)) {
		panic("affected rows is not equal to transactions size")
	}

	for _, t := range transactions {
		t.Response <- nil
	}

	n.logger.Sugar().Debugf("bulk insert %d transactions done", len(transactions))
}
