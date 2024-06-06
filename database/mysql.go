package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type MysqlDriver struct {
	conn *sql.DB
	config Config
}

func (m MysqlDriver) Open(cfg Config) (Driver, error) {
	config := mysql.Config{
		User:   cfg.Username,
		Passwd: cfg.Password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DBName: cfg.Database,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	d := MysqlDriver{
		conn: db,
		config: cfg,
	}

	return d, nil
}

func (m MysqlDriver) Close() error {
	return nil
}

func (m MysqlDriver) Run(query string) error {
	_, err := m.conn.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (m MysqlDriver) Wipe() error {
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

func (m MysqlDriver) GetConnection() *sql.DB {
	return m.conn
}