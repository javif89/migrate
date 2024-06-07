package database

import "database/sql"

type TestingDriver struct {
	Queries []string
}

func (t TestingDriver) Open(cfg Config) (Driver, error) {
	d := TestingDriver{
		Queries: []string{},
	}

	return d, nil
}

func (t TestingDriver) Close() error {
	return nil
}

func (t TestingDriver) Run(query string) error {
	t.addQuery(query)

	return nil
}

func (t TestingDriver) Wipe() error {
	t.wipeQueries()

	return nil
}

func (t TestingDriver) GetConnection() *sql.DB {
	return &sql.DB{}
}

func (t TestingDriver) CreateMigrationsTable() error {
	// m.conn.Exec()

	return nil
}

func (t *TestingDriver) addQuery(q string) {
	t.Queries = append(t.Queries, q)
}

func (t *TestingDriver) wipeQueries() {
	t.Queries = []string{}
}