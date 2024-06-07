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
	query := `SELECT name FROM sqlite_master WHERE type = 'table';`
	tables, err := m.conn.Query(query)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := tables.Close(); errClose != nil {
			fmt.Println("Err close ", errClose)
		}
	}()

	tableNames := make([]string, 0)
	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return err
		}
		if len(tableName) > 0 {
			tableNames = append(tableNames, tableName)
		}
	}
	if err := tables.Err(); err != nil {
		return err
	}

	if len(tableNames) > 0 {
		for _, t := range tableNames {
			query := "DROP TABLE " + t
			_, err = m.conn.Exec(query)
			if err != nil {
				return err
			}
		}
		query := "VACUUM"
		_, err = m.conn.Query(query)
		if err != nil {
			return err
		}
	}

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