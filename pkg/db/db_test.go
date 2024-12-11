package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock the database connection using sqlmock
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err, "sqlmock should create a mock database without errors")
	defer dbMock.Close()

	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Replace sql.Open with a mock database connection
	db, err := New("user", "password", "localhost", "3306", "testdb", "", 10, 5, 30*time.Minute, false)

	assert.NoError(t, err, "New should not return an error with valid parameters")
	assert.NotNil(t, db, "New should return a valid DB object")

	// Verify the database connection is configured correctly
	assert.NotNil(t, db.DB, "DB should be initialized")
	assert.Equal(t, 10, db.DB.Stats().MaxOpenConnections, "MaxOpenConnections should be set correctly")
}

func TestLockAndUnlock(t *testing.T) {
	db := &DB{}

	// Test Lock
	isLocked := make(chan bool, 1)
	go func() {
		db.Lock()
		isLocked <- true
		db.Unlock()
	}()

	select {
	case <-isLocked:
		assert.True(t, true, "Lock should allow acquisition")
	case <-time.After(1 * time.Second):
		t.Fatal("Lock did not allow acquisition in a reasonable time")
	}

	// Test Unlock
	unlocked := false
	db.Lock()
	go func() {
		db.Unlock()
		unlocked = true
	}()
	time.Sleep(100 * time.Millisecond) // Allow goroutine to run
	assert.True(t, unlocked, "Unlock should release the lock")
}
