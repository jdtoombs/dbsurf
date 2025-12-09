package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/microsoft/go-mssqldb"
)

func DetectDBType(connString string) string {
	if strings.HasPrefix(connString, "postgres") {
		return "postgres"
	}
	if strings.HasPrefix(connString, "sqlserver") {
		return "sqlserver"
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
	case "sqlserver":
		query = "SELECT name FROM sys.databases"
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

func UseDatabase(db *sql.DB, name, dbType string) error {
	var query string
	switch dbType {
	case "sqlserver":
		query = fmt.Sprintf("USE [%s]", name)
	default:
		query = "USE " + name
	}
	_, err := db.Exec(query)
	return err
}

func ListTables(db *sql.DB, dbName, dbType string) ([]string, error) {
	var query string
	var includeSchema bool
	switch dbType {
	case "mysql":
		query = "SHOW TABLES"
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	case "sqlserver":
		query = fmt.Sprintf("SELECT TABLE_SCHEMA, TABLE_NAME FROM [%s].INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'", dbName)
		includeSchema = true
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		if includeSchema {
			var schema, name string
			rows.Scan(&schema, &name)
			tables = append(tables, fmt.Sprintf("[%s].[%s]", schema, name))
		} else {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
	}
	return tables, nil
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
