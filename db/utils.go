package db

import "strings"

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

// PrependUseDatabase prepends USE [database]; to a query for SQL Server.
// For other database types, returns the query unchanged.
func PrependUseDatabase(query, database, dbType string) string {
	if dbType != "sqlserver" {
		return query
	}
	return "USE [" + database + "]; " + query
}
