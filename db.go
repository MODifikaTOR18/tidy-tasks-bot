package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/lib/pq"
)

func ConnectDB(DBInfo DBInfo) *sql.DB {
	var err error
	connectionString := "postgresql://" + DBInfo.DBUser + ":" + DBInfo.DBPassword + "@postgres:5432/tidytasks?sslmode=disable"

	connector, err := pq.NewConnector(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	db := sql.OpenDB(connector)

	return db
}

func ExecQuery(DBInfo DBInfo, query string) (sql.Result, error) {
	db := ConnectDB(DBInfo)

	result, err := db.Exec(query)
	log.Printf("Executing query: %v", query)

	return result, err
}

func InitDB(DBInfo DBInfo) {
	sqlQueryRaw, err := os.ReadFile("db-init.sql")
	if err != nil {
		log.Fatalf("Failed to read sql init dump: %v", err)
	}

	sqlQuery := string(sqlQueryRaw)
	execResult, err := ExecQuery(DBInfo, sqlQuery)
	if err != nil {
		log.Fatalf("Failed to initiate database: %v", err)
	}

	execRows, err := execResult.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get number of rows affected by SQL query: %v", err)
	}
	log.Printf("Rows affected by DB initialization: %v", execRows)
}

func CreateUser(DBInfo DBInfo, userName string, userID string) int {
	sqlQuery := "INSERT INTO users (telegram_id, name) VALUES ('" + userID + "','" + userName + "');"
	execResult, err := ExecQuery(DBInfo, sqlQuery)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 1
		}
	}

	execRows, err := execResult.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get number of rows affected by SQL query: %v", err)
	}
	log.Printf("Rows affected by creating a new user: %v", execRows)
	return 0
}

func CreateNewTask(DBInfo DBInfo, userID string, description string, scheduledTime string, isRecurring string, interval string) int64 {
	sqlResult, err := ExecQuery(DBInfo, "INSERT INTO tasks (user_id, description, scheduled_time, is_recurring, interval) VALUES ('"+userID+"','"+description+"','"+scheduledTime+"','"+isRecurring+"','"+interval+"')")
	if err != nil {
		log.Fatalf("Failed to create new task: %v", err)
	}

	execRows, err := sqlResult.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get number of rows affected by SQL query: %v", err)
	}
	log.Printf("Rows affected by creating a new user: %v", execRows)
	return execRows
}
