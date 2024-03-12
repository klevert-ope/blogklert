package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}

	// Set maximum number of open connections
	DB.SetMaxOpenConns(10)

	// Set maximum number of idle connections
	DB.SetMaxIdleConns(5)

	// Set maximum amount of time a connection can be idle before it is closed
	DB.SetConnMaxIdleTime(30 * time.Minute)

	// Check if the connection is still alive
	err = DB.Ping()
	if err != nil {
		return err
	}

	log.Println("Database connection initialized successfully.")
	return nil
}

func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed successfully.")
		}
	}
}
