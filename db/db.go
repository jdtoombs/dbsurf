package db

import (
	"database/sql"
	"fmt"
	"strings"
	"github.com/go-sql-driver/mysql"
)

func DetectDBType(connString string) string {
	if strings.HasPrefix(connString, "postgres") {
		return "postgres"
	}
	return "mysql"
}

func Connect(connString string) (*sql.DB, error) {
	dbType := DetectDBType(connString)

	db, err := sql.Open(dbType, connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return db, nil
}

func ListDatabases(db *sql.DB, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgres":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		databases = append(databases, name)
	}
	return databases, nil
}

func UseDatabase(db *sql.DB, name string) error {
	_, err := db.Exec("USE " + name)
	return err
}

type QueryResult struct {
	Columns []string
	Rows    [][]string
}

func RunQuery(db *sql.DB, query string) (*QueryResult, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	result := &QueryResult{Columns: columns}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		rows.Scan(pointers...)

		row := make([]string, len(columns))
		for i, v := range values {
			if v == nil {
				row[i] = "NULL"
			} else if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}
