package db

import (
	"testing"
)

func TestValidateConnectionString_Empty(t *testing.T) {
	tests := []struct {
		name    string
		conn    string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionString(tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConnectionString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConnectionString_Postgres(t *testing.T) {
	tests := []struct {
		name     string
		conn     string
		wantErr  bool
		errField string
	}{
		{
			name:    "valid full",
			conn:    "postgres://user:pass@localhost:5432/mydb",
			wantErr: false,
		},
		{
			name:    "valid no port",
			conn:    "postgres://user:pass@localhost/mydb",
			wantErr: false,
		},
		{
			name:    "valid with params",
			conn:    "postgres://user:pass@localhost/db?sslmode=disable",
			wantErr: false,
		},
		{
			name:    "valid postgresql scheme",
			conn:    "postgresql://user:pass@localhost/db",
			wantErr: false,
		},
		{
			name:    "valid no password",
			conn:    "postgres://user@localhost/db",
			wantErr: false,
		},
		{
			name:     "missing host",
			conn:     "postgres:///mydb",
			wantErr:  true,
			errField: "host",
		},
		{
			name:     "missing user",
			conn:     "postgres://localhost:5432/mydb",
			wantErr:  true,
			errField: "user",
		},
		{
			name:     "invalid port letters",
			conn:     "postgres://user:pass@localhost:abc/db",
			wantErr:  true,
			errField: "format",
		},
		{
			name:     "port out of range",
			conn:     "postgres://user:pass@localhost:99999/db",
			wantErr:  true,
			errField: "port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionString(tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConnectionString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errField != "" {
				if ve, ok := err.(*ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("ValidationError.Field = %v, want %v", ve.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestValidateConnectionString_SQLServer(t *testing.T) {
	tests := []struct {
		name     string
		conn     string
		wantErr  bool
		errField string
	}{
		{
			name:    "valid full",
			conn:    "sqlserver://user:pass@localhost:1433?database=mydb",
			wantErr: false,
		},
		{
			name:    "valid no port",
			conn:    "sqlserver://user:pass@localhost?database=mydb",
			wantErr: false,
		},
		{
			name:    "valid no database param",
			conn:    "sqlserver://user:pass@localhost:1433",
			wantErr: false,
		},
		{
			name:     "missing host",
			conn:     "sqlserver://?database=mydb",
			wantErr:  true,
			errField: "host",
		},
		{
			name:     "missing user",
			conn:     "sqlserver://localhost:1433?database=mydb",
			wantErr:  true,
			errField: "user",
		},
		{
			name:     "invalid port letters",
			conn:     "sqlserver://user:pass@localhost:abc?database=db",
			wantErr:  true,
			errField: "format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionString(tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConnectionString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errField != "" {
				if ve, ok := err.(*ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("ValidationError.Field = %v, want %v", ve.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestValidateConnectionString_MySQL(t *testing.T) {
	tests := []struct {
		name     string
		conn     string
		wantErr  bool
		errField string
	}{
		{
			name:    "valid full",
			conn:    "user:pass@tcp(localhost:3306)/mydb",
			wantErr: false,
		},
		{
			name:    "valid no database",
			conn:    "user:pass@tcp(localhost:3306)/",
			wantErr: false,
		},
		{
			name:    "valid no password",
			conn:    "user@tcp(localhost:3306)/mydb",
			wantErr: false,
		},
		{
			name:    "valid no port",
			conn:    "user:pass@tcp(localhost)/mydb",
			wantErr: false,
		},
		{
			name:     "missing tcp wrapper",
			conn:     "user:pass@localhost:3306/mydb",
			wantErr:  true,
			errField: "format",
		},
		{
			name:     "empty host",
			conn:     "user:pass@tcp(:3306)/mydb",
			wantErr:  true,
			errField: "host",
		},
		{
			name:     "invalid port",
			conn:     "user:pass@tcp(localhost:abc)/mydb",
			wantErr:  true,
			errField: "port",
		},
		{
			name:     "completely invalid",
			conn:     "not-a-valid-dsn",
			wantErr:  true,
			errField: "format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionString(tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConnectionString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errField != "" {
				if ve, ok := err.(*ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("ValidationError.Field = %v, want %v", ve.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		port  string
		valid bool
	}{
		{"1", true},
		{"80", true},
		{"443", true},
		{"3306", true},
		{"5432", true},
		{"65535", true},
		{"0", false},
		{"65536", false},
		{"abc", false},
		{"-1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			if got := isValidPort(tt.port); got != tt.valid {
				t.Errorf("isValidPort(%q) = %v, want %v", tt.port, got, tt.valid)
			}
		})
	}
}
