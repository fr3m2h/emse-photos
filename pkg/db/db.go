package db

import (
	"database/sql"
	"fmt"
	"photos/pkg/db/query"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB is a wrapper around the standard sql.DB struct, adding a mutex for thread-safe
// operations and an embedded Queries struct for interacting with the database.
type DB struct {
	*sql.DB                   // The underlying SQL database connection.
	mux            sync.Mutex // Mutex to provide thread-safe access.
	*query.Queries            // Query methods for interacting with the database.
}

// New creates and configures a new MySQL database connection.
//
// Parameters:
//   - username: The username for connecting to the database.
//   - password: The password for the database user.
//   - host: The hostname or IP address of the database server.
//   - port: The port on which the database server is listening.
//   - dbName: The name of the database to connect to.
//   - cert: Path to the TLS certificate file (unused in this function but can be extended for secure connections).
//   - maxOpenConns: Maximum number of open connections to the database.
//   - maxIdleConns: Maximum number of idle connections in the pool.
//   - connMaxLifetime: Maximum lifetime of a connection.
//   - useTLS: Boolean indicating whether to use TLS for the connection.
//
// Returns:
//   - *DB: A pointer to the initialized DB struct.
//   - error: An error if the connection could not be established.
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

// Lock acquires the mutex lock for thread-safe operations on the DB object.
func (db *DB) Lock() {
	db.mux.Lock()
}

// Unlock releases the mutex lock for thread-safe operations on the DB object.
func (db *DB) Unlock() {
	db.mux.Unlock()
}
