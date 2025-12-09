package db

import (
	"testing"
)

func TestDetectDBType_Postgres(t *testing.T) {
	tests := []struct {
		connString string
		want       string
	}{
		{"postgres://localhost/test", "postgres"},
		{"postgres://user:pass@localhost:5432/mydb", "postgres"},
		{"postgresql://localhost/test", "postgres"}, // Note: this will return mysql since it checks HasPrefix("postgres")
	}

	for _, tt := range tests {
		got := DetectDBType(tt.connString)
		if got != tt.want {
			t.Errorf("DetectDBType(%q) = %q, want %q", tt.connString, got, tt.want)
		}
	}
}

func TestDetectDBType_SQLServer(t *testing.T) {
	tests := []struct {
		connString string
		want       string
	}{
		{"sqlserver://localhost", "sqlserver"},
		{"sqlserver://user:pass@localhost:1433?database=mydb", "sqlserver"},
	}

	for _, tt := range tests {
		got := DetectDBType(tt.connString)
		if got != tt.want {
			t.Errorf("DetectDBType(%q) = %q, want %q", tt.connString, got, tt.want)
		}
	}
}

func TestDetectDBType_MySQL(t *testing.T) {
	tests := []struct {
		connString string
		want       string
	}{
		{"user:pass@tcp(localhost:3306)/mydb", "mysql"},
		{"root@tcp(127.0.0.1:3306)/test", "mysql"},
		{"mysql://localhost/test", "mysql"}, // mysql:// prefix defaults to mysql
		{"anything-else", "mysql"},          // default is mysql
	}

	for _, tt := range tests {
		got := DetectDBType(tt.connString)
		if got != tt.want {
			t.Errorf("DetectDBType(%q) = %q, want %q", tt.connString, got, tt.want)
		}
	}
}

func TestConnect_InvalidDSN(t *testing.T) {
	// Test with clearly invalid connection strings
	invalidDSNs := []string{
		"postgres://invalid:5432/nonexistent",
		"not-a-valid-connection-string",
	}

	for _, dsn := range invalidDSNs {
		_, err := Connect(dsn)
		if err == nil {
			t.Errorf("Connect(%q) should return error for invalid DSN", dsn)
		}
	}
}

func TestQueryResult_Structure(t *testing.T) {
	// Test that QueryResult can hold expected data
	result := &QueryResult{
		Columns: []string{"id", "name", "email"},
		Rows: [][]string{
			{"1", "Alice", "alice@example.com"},
			{"2", "Bob", "bob@example.com"},
		},
	}

	if len(result.Columns) != 3 {
		t.Errorf("QueryResult.Columns = %d, want 3", len(result.Columns))
	}
	if len(result.Rows) != 2 {
		t.Errorf("QueryResult.Rows = %d, want 2", len(result.Rows))
	}
	if result.Rows[0][1] != "Alice" {
		t.Errorf("QueryResult.Rows[0][1] = %q, want %q", result.Rows[0][1], "Alice")
	}
}
