package db

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var tiDB *TiDb

type TiDb struct {
	conn       []*sql.DB
	roundRobin int
	dbConfig   databaseConfig
}

func New(logger *zap.Logger) (*TiDb, error) {
	if tiDB != nil {
		return tiDB, nil
	}
	tiDB = &TiDb{}
	dbConfig, err := ReadDatabaseConfig(logger)
	if err != nil {
		return nil, fmt.Errorf("error reading database config: %w", err)
	}
	tiDB.dbConfig = dbConfig
	logger.Info("Connecting to database...")
	if dbConfig.TLS == "tidb" {
		logger.Sugar().Infof("Connecting to %s", dbConfig.Address[0])
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile("./ca.pem")
		if err != nil {
			logger.Sugar().Errorf("Failed to read client certificate authority: %v", err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			logger.Fatal("Failed to append PEM.")
		}
		mysql.RegisterTLSConfig("tidb", &tls.Config{
			RootCAs:    rootCertPool,
			MinVersion: tls.VersionTLS12,
			ServerName: "tidb.zpdqjj100n16.clusters.tidb-cloud.com",
		})

		dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=%s&parseTime=True&tidb_skip_isolation_level_check=1",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Address[0],
			dbConfig.DatabaseName,
			dbConfig.TLS,
		)

		conn, err := sql.Open("mysql", dataSourceName)
		if err != nil {
			log.Fatal("failed to connect database", err)
		}
		conn.SetMaxOpenConns(dbConfig.MaxOpenConns)
		conn.SetMaxIdleConns(dbConfig.MaxIdleConns)
		conn.SetConnMaxIdleTime(dbConfig.ConnMaxIdle)
		conn.SetConnMaxLifetime(dbConfig.ConnMaxLife)

		err = conn.Ping()
		if err != nil {
			return nil, fmt.Errorf("error pinging database: %w", err)
		}

		tiDB.conn = append(tiDB.conn, conn)

	} else {
		for _, addr := range dbConfig.Address {
			dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=%s&parseTime=True&tidb_skip_isolation_level_check=1",
				dbConfig.Username,
				dbConfig.Password,
				addr,
				dbConfig.DatabaseName,
				dbConfig.TLS,
			)
			logger.Sugar().Infof("Connecting to %s", dataSourceName)
			conn, err := sql.Open("mysql", dataSourceName)
			if err != nil {
				return nil, fmt.Errorf("error opening database connection: %w", err)
			}
			conn.SetMaxOpenConns(dbConfig.MaxOpenConns)
			conn.SetMaxIdleConns(dbConfig.MaxIdleConns)
			conn.SetConnMaxIdleTime(dbConfig.ConnMaxIdle)
			conn.SetConnMaxLifetime(dbConfig.ConnMaxLife)

			err = conn.Ping()
			if err != nil {
				return nil, fmt.Errorf("error pinging database: %w", err)
			}
			tiDB.conn = append(tiDB.conn, conn)
		}
	}
	logger.Info("Connect to database successfully")
	return tiDB, nil
}

func (db *TiDb) GetConnRoundRobin() *sql.DB {
	db.roundRobin = (db.roundRobin + 1) % len(db.conn)
	return db.conn[db.roundRobin]
}

func (db *TiDb) Close() error {
	for _, conn := range db.conn {
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("error closing database connection: %w", err)
		}
	}
	return nil
}
