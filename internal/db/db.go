package db

import (
	"database/sql"
	"fmt"
	"photos/internal/db/query"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB is a wrapper around the standard sql.DB struct. It includes a mutex for ensuring
// thread-safe operations and an embedded Queries struct for database interaction.
// This struct simplifies database access by combining connection management and query methods.
type DB struct {
	*sql.DB                   // The underlying SQL database connection.
	mux            sync.Mutex // A mutex to ensure thread-safe access.
	*query.Queries            // Query methods for interacting with the database.
}

// New creates and configures a new MySQL database connection. It establishes a connection
// using the provided credentials and database details. The function configures connection pooling
// by setting limits for open connections, idle connections, and the maximum lifetime of a connection.
// It also supports enabling or disabling TLS for secure database connections.
func New(username, password, host, port, dbName, cert string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration, useTLS bool) (*DB, error) {
	mysqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=%t&loc=UTC&parseTime=true", username, password, host, port, dbName, useTLS))
	if err != nil {
		return nil, err
	}
	mysqlDB.SetConnMaxLifetime(connMaxLifetime)
	mysqlDB.SetMaxOpenConns(maxOpenConns)
	mysqlDB.SetMaxIdleConns(maxIdleConns)

	db := &DB{DB: mysqlDB, mux: sync.Mutex{}, Queries: query.New(mysqlDB)}
	return db, nil
}

// Lock acquires the mutex lock for the DB object. This ensures that concurrent
// operations on the database connection are thread-safe.
func (db *DB) Lock() {
	db.mux.Lock()
}

// Unlock releases the mutex lock for the DB object. This allows other threads
// to safely perform operations on the database connection.
func (db *DB) Unlock() {
	db.mux.Unlock()
}
