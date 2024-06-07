package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/javif89/migrate/database"
)

func TestCreateMigration(t *testing.T) {
	d := t.TempDir()
	m := New(d, database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	m.CreateMigration("test_migration")

	f, err := os.ReadDir(d)

	if err != nil {
		t.Fatalf("Failed reading temp directory")
	}

	found := false
	var foundFilename string // For checking file contents

	for _, file := range f {
		if strings.Contains(file.Name(), "test_migration.sql") {
			found = true
			foundFilename = filepath.Join(d, file.Name())
		}
	}

	if !found {
		t.Fatalf("Migration file was not created")
	}

	// Check that the contents are correct
	c, _ := os.ReadFile(foundFilename)

	if !(string(c) == "-- UP --\n\n-- DOWN --") {
		t.Errorf("Migration content is not correct")
	}
}

func TestMigrate(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	conn := m.driver.GetConnection()

	m.CreateMigration("test_migration")
	m.Migrate()

	// Assert that the migrations table was created
	r := conn.QueryRow("select name from sqlite_master where type = 'table' and name = 'migrations'")

	var name string
	if err := r.Scan(&name); err != nil && err == sql.ErrNoRows {
		fmt.Println(name)
		t.Fatalf("Migrations table was not created")
	}

	// Assert that the migration was run
	r = conn.QueryRow("select batch from migrations where migration like '%test_migration%'")
	if err := r.Scan(); err != nil && err == sql.ErrNoRows {
		t.Fatalf("Migration was not added to the migrations table")
	}
}

func TestMigrateBatches(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	conn := m.driver.GetConnection()

	m.CreateMigration("test_migration")
	m.Migrate()

	// Assert that the migration has the correct batch
	r := conn.QueryRow("select batch from migrations where migration like '%test_migration%'")
	var b int
	r.Scan(&b)

	if b != 1 {
		t.Errorf("Incorrect batch number on the first migration %d", b)
	}

	m.CreateMigration("next_migration")
	m.Migrate()

	// Assert that the migration has the correct batch
	r = conn.QueryRow("select batch from migrations where migration like '%next_migration%'")
	r.Scan(&b)

	if b != 2 {
		t.Errorf("Incorrect batch number on the second migration %d", b)
	}

	// Assert that only two migrations have been run
	r = conn.QueryRow("select count(*) from migrations")
	var count int
	r.Scan(&count)

	if count != 2 {
		t.Errorf("More than two migrations were added to the table")
	}
}

func TestRollback(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	conn := m.driver.GetConnection()

	m.CreateMigration("test_migration")
	m.Migrate()

	m.CreateMigration("next_migration")
	m.Migrate()

	// Assert that both migrations ran
	r := conn.QueryRow("select count(*) from migrations")
	var count int
	r.Scan(&count)

	if count != 2 {
		t.Errorf("Not all migrations ran")
	}

	m.Rollback()

	// Assert that only the last batch was rolled back
	r = conn.QueryRow("select count(*) from migrations")
	r.Scan(&count)

	if count != 1 {
		t.Errorf("Rollback did more than the last batch")
	}
}

func TestFresh(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	conn := m.driver.GetConnection()

	m.CreateMigration("test_migration")
	m.Migrate()

	m.CreateMigration("next_migration")
	m.Migrate()

	// Check that we have two different batches
	var count int
	r := conn.QueryRow("select count(distinct batch) from migrations")
	r.Scan(&count)

	if count != 2 {
		t.Errorf("Batch numbers are incorrect. We have %d batch(es)", count)
	}

	m.Fresh()

	m = New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})
	conn = m.driver.GetConnection()

	// Check that we have only one batch since it should have
	// wiped the DB and re-run all migrations
	r = conn.QueryRow("select count(distinct batch) from migrations")
	r.Scan(&count)

	if count != 1 {
		t.Errorf("Batch numbers are incorrect after migrate fresh. We have %d batch(es)", count)
	}
}

func TestGetMigrations(t *testing.T) {
	d := t.TempDir()
	mgf := filepath.Join(d, "migrations") // Migrations folder path
	os.MkdirAll(mgf, os.ModePerm)
	m := New(mgf, database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	if _, err := m.GetMigrations(); err != ErrNoMigrations {
		t.Error("Not throwing an error when there are no migration files")
	}

	m.CreateMigration("first")
	m.CreateMigration("second")

	mg, _ := m.GetMigrations()

	if len(mg) != 2 {
		t.Errorf("Not all migrations were loaded")
	}

	if !strings.Contains(mg[0].Name(), "first") {
		t.Errorf("Migrations were read in the wrong order")
	}
}

func TestGetMigrationsReverse(t *testing.T) {
	d := t.TempDir()
	mgf := filepath.Join(d, "migrations") // Migrations folder path
	os.MkdirAll(mgf, os.ModePerm)
	m := New(mgf, database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	m.CreateMigration("first")
	m.CreateMigration("second")

	mg, _ := m.GetMigrationsReverse()

	if len(mg) != 2 {
		t.Errorf("Not all migrations were loaded")
	}

	if !strings.Contains(mg[0].Name(), "second") {
		t.Errorf("Migrations were read in the wrong order")
	}
}

func TestGetUnexecutedMigrations(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	m.CreateMigration("done")
	m.Migrate()
	m.CreateMigration("not_done")

	mg, _ := m.GetUnexecutedMigrations()

	if len(mg) != 1 {
		t.Errorf("Incorrect number of unexecuted migrations: %d", len(mg))
	}

	if !strings.Contains(mg[0].Name(), "not_done") {
		t.Errorf("Incorrect unexecuted migration: %s", mg[0].Name())
	}
}

func TestGetExistingMigrations(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	m.CreateMigration("done")
	m.Migrate()
	m.CreateMigration("not_done")

	mg := m.GetExistingMigrations()

	if len(mg) != 1 {
		t.Errorf("Incorrect number of existing migrations: %d", len(mg))
	}

	if !strings.Contains(mg[0], "done") {
		t.Errorf("Incorrect existing migration: %s", mg[0])
	}
}

func TestGetMigrationsInBatch(t *testing.T) {
	d := t.TempDir()
	m := New(filepath.Join(d, "migrations"), database.DriverSqlite, database.Config{
		Database: filepath.Join(d, "testdb.sqlite"),
	})

	m.CreateMigration("first")
	m.Migrate()
	m.CreateMigration("second")
	m.Migrate()

	mg := m.GetMigrationsInBatch(1)

	if len(mg) != 1 {
		t.Errorf("Incorrect number of migrations in batch 1: %d", len(mg))
	}

	mg = m.GetMigrationsInBatch(2)

	if len(mg) != 1 {
		t.Errorf("Incorrect number of migrations in batch 2: %d", len(mg))
	}
}