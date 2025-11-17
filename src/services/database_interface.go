package services

import (
	"database/sql"
)

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetDB() *sql.DB
	GetPath() string
}