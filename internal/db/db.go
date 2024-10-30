package db

import (
	"database/sql"
	"fmt"
	"photos/internal/db/query"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
	mux sync.Mutex
	*query.Queries
}

// New creates a new database connection
func New(username, password, host, port, dbName, cert string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration, useTLS bool) (*DB, error) {
	mysqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=%t&loc=UTC&parseTime=true", username, password, host, port, dbName, useTLS))
	if err != nil {
		return nil, err
	}
	mysqlDB.SetConnMaxLifetime(connMaxLifetime)
	mysqlDB.SetMaxOpenConns(maxOpenConns)
	mysqlDB.SetMaxIdleConns(maxIdleConns)

	return &DB{DB: mysqlDB, mux: sync.Mutex{}, Queries: query.New(mysqlDB)}, nil
}

func (db *DB) Lock() {
	db.mux.Lock()
}

func (db *DB) Unlock() {
	db.mux.Unlock()
}
