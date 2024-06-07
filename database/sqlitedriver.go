package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type SQLiteDriver struct {
	conn *sql.DB
	config Config
}

func (m SQLiteDriver) Open(cfg Config) (Driver, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_txlock=immediate", cfg.Database))
	if err != nil {
		return nil, err
	}

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	d := SQLiteDriver{
		conn: db,
		config: cfg,
	}

	return d, nil
}

func (m SQLiteDriver) Close() error {
	return nil
}

func (m SQLiteDriver) Run(query string) error {
	_, err := m.conn.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (m SQLiteDriver) Wipe() error {
	// Get all the table drop commands
	q := fmt.Sprintf(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = '%s';
	`, m.config.Database)

	rows, err := m.conn.Query(q)

	if err != nil {
		return err
	}

	defer rows.Close()

	// Disable foreign keys
	m.conn.Exec("SET FOREIGN_KEY_CHECKS = 0;")

	// Delete all tables
	for rows.Next() {
		var t string
		rows.Scan(&t)
		m.conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", t))
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	// Enable foreign keys
	m.conn.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	return nil
}

func (m SQLiteDriver) CreateMigrationsTable() error {
	_, err := m.conn.Exec(`
		create table if not exists migrations (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			migration varchar(255),
			batch int
		)
	`)

	return err
}

func (m SQLiteDriver) GetConnection() *sql.DB {
	return m.conn
}