// utils.go provides SQL Server-specific utilities for handling schema-qualified
// table names and query formatting.
package db

import (
	"strings"
)

// CleanTableName removes schema prefix and brackets from table names.
// For example: "[dbo].[users]" -> "users", "dbo.users" -> "users"
func CleanTableName(tableName, dbType string) string {
	if dbType != "sqlserver" {
		return tableName
	}

	cleanName := tableName
	if idx := strings.LastIndex(tableName, "."); idx != -1 {
		cleanName = tableName[idx+1:]
	}
	return strings.Trim(cleanName, "[]\"` ")
}

// ExtractSchema extracts the schema from a table name.
// For example: "[dbo].[users]" -> "dbo", "users" -> "dbo" (default)
func ExtractSchema(tableName, dbType string) string {
	if dbType != "sqlserver" {
		return ""
	}

	if idx := strings.LastIndex(tableName, "."); idx != -1 {
		schema := tableName[:idx]
		return strings.Trim(schema, "[]\"` ")
	}
	return "dbo" // Default schema for SQL Server
}

// PrependUseDatabase prepends USE [database]; to a query for SQL Server.
// For other database types, returns the query unchanged.
func PrependUseDatabase(query, database, dbType string) string {
	if dbType != "sqlserver" {
		return query
	}
	return "USE [" + database + "]; " + query
}

// FormatTableName formats a table name for use in SQL queries.
// For SQL Server, returns [schema].[table]. For others, returns as-is.
func FormatTableName(tableName, dbType string) string {
	if dbType != "sqlserver" {
		return tableName
	}

	// Handle schema.table format
	if idx := strings.Index(tableName, "."); idx != -1 {
		schema := strings.Trim(tableName[:idx], "[]")
		table := strings.Trim(tableName[idx+1:], "[]")
		return "[" + schema + "].[" + table + "]"
	}
	// No schema, just wrap table name
	return "[" + strings.Trim(tableName, "[]") + "]"
}
