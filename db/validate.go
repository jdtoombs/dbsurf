// validate.go provides connection string validation for different database types.
// It validates URL format, required fields, and port ranges before attempting
// to connect.
package db

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidateConnectionString(connString string) error {
	connString = strings.TrimSpace(connString)
	if connString == "" {
		return &ValidationError{Field: "connection", Message: "connection string cannot be empty"}
	}

	dbType := DetectDBType(connString)

	switch dbType {
	case "postgres":
		return validatePostgres(connString)
	case "sqlserver":
		return validateSQLServer(connString)
	default:
		return validateMySQL(connString)
	}
}

func validatePostgres(connString string) error {
	u, err := url.Parse(connString)
	if err != nil {
		return &ValidationError{Field: "format", Message: "invalid URL format"}
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return &ValidationError{Field: "scheme", Message: "PostgreSQL connection must start with postgres:// or postgresql://"}
	}

	if u.Host == "" {
		return &ValidationError{Field: "host", Message: "host is required (e.g., postgres://user:pass@localhost:5432/db)"}
	}

	host := u.Hostname()
	if host == "" {
		return &ValidationError{Field: "host", Message: "hostname cannot be empty"}
	}

	if port := u.Port(); port != "" {
		if !isValidPort(port) {
			return &ValidationError{Field: "port", Message: fmt.Sprintf("invalid port: %s", port)}
		}
	}

	if u.User == nil || u.User.Username() == "" {
		return &ValidationError{Field: "user", Message: "username is required (e.g., postgres://user:pass@host/db)"}
	}

	return nil
}

func validateSQLServer(connString string) error {
	u, err := url.Parse(connString)
	if err != nil {
		return &ValidationError{Field: "format", Message: "invalid URL format"}
	}

	if u.Scheme != "sqlserver" {
		return &ValidationError{Field: "scheme", Message: "SQL Server connection must start with sqlserver://"}
	}

	if u.Host == "" {
		return &ValidationError{Field: "host", Message: "host is required (e.g., sqlserver://user:pass@localhost:1433?database=db)"}
	}

	host := u.Hostname()
	if host == "" {
		return &ValidationError{Field: "host", Message: "hostname cannot be empty"}
	}

	if port := u.Port(); port != "" {
		if !isValidPort(port) {
			return &ValidationError{Field: "port", Message: fmt.Sprintf("invalid port: %s", port)}
		}
	}

	if u.User == nil || u.User.Username() == "" {
		return &ValidationError{Field: "user", Message: "username is required (e.g., sqlserver://user:pass@host?database=db)"}
	}

	return nil
}

func validateMySQL(connString string) error {
	// MySQL DSN format: user:pass@tcp(host:port)/dbname
	// or: user@tcp(host:port)/dbname (no password)
	tcpRegex := regexp.MustCompile(`^([^:@]+)(:[^@]*)?@tcp\(([^)]+)\)(/.*)?$`)
	matches := tcpRegex.FindStringSubmatch(connString)

	if matches == nil {
		return &ValidationError{
			Field: "format",
			Message: "invalid format. expected:\n" +
				"  postgres://user:pass@host:5432/db\n" +
				"  sqlserver://user:pass@host:1433?database=db\n" +
				"  user:pass@tcp(host:3306)/db",
		}
	}

	user := matches[1]
	if user == "" {
		return &ValidationError{Field: "user", Message: "username is required"}
	}

	hostPort := matches[3]
	if hostPort == "" {
		return &ValidationError{Field: "host", Message: "host is required inside tcp()"}
	}

	parts := strings.Split(hostPort, ":")
	if parts[0] == "" {
		return &ValidationError{Field: "host", Message: "hostname cannot be empty"}
	}

	if len(parts) > 1 && parts[1] != "" {
		if !isValidPort(parts[1]) {
			return &ValidationError{Field: "port", Message: fmt.Sprintf("invalid port: %s", parts[1])}
		}
	}

	return nil
}

func isValidPort(port string) bool {
	var p int
	_, err := fmt.Sscanf(port, "%d", &p)
	return err == nil && p > 0 && p <= 65535
}
