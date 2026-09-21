package sync

import (
	"fmt"
	"sync"
)

type DatabaseConnection struct {
	URL string
}

var (
	dbInstance *DatabaseConnection
	once       sync.Once
)

func GetDBInstance() *DatabaseConnection {
	once.Do(func() {
		// This block executes EXACTLY once across the life of the process.
		fmt.Println("Initializing expensive database connection pool...")
		dbInstance = &DatabaseConnection{URL: "postgres://localhost:5432/db"}
	})
	return dbInstance
}