package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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

	// Configure pool for single-user TUI
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

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

func GetPrimaryKey(db *sql.DB, dbName, tableName, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf(`
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s' AND CONSTRAINT_NAME = 'PRIMARY'
			ORDER BY ORDINAL_POSITION`, dbName, tableName)
	case "postgres":
		query = fmt.Sprintf(`
			SELECT a.attname
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			WHERE i.indrelid = '%s'::regclass AND i.indisprimary
			ORDER BY array_position(i.indkey, a.attnum)`, tableName)
	case "sqlserver":
		query = fmt.Sprintf(`
			SELECT COLUMN_NAME
			FROM [%s].INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE OBJECTPROPERTY(OBJECT_ID(CONSTRAINT_SCHEMA + '.' + CONSTRAINT_NAME), 'IsPrimaryKey') = 1
			AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION`, dbName, tableName)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		rows.Scan(&col)
		columns = append(columns, col)
	}
	return columns, nil
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
