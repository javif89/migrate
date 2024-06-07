package database

import "database/sql"

type Driver interface {
	Open(cfg Config) (Driver, error)
	// Close closes the underlying database instance managed by the driver.
	// Migrate will call this function only once per instance.
	Close() error

	// Get the underlying connection. We can use this for
	// queries we might need to do such as getting the
	// next batch number
	GetConnection() *sql.DB

	// Run applies a migration to the database. migration is guaranteed to be not nil.
	Run(query string) error

	// Wipe deletes everything in the database.
	Wipe() error

	CreateMigrationsTable() error
}