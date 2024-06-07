package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/javif89/migrate/database"
)

var ErrNoMigrations error = errors.New("no migrations") 
var ErrNoMigrationsToRun error = errors.New("nothing to migrate")

type Migrations struct {
	path string
	driver database.Driver
}

func New(path string, driver database.DriverName, cfg database.Config) *Migrations {
	d, err := database.GetDriver(driver, cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &Migrations{
		path: path,
		driver: d,
	}
}

func (m *Migrations) CreateMigration(name string) {
	n := getMigrationFileName(name)
	filename := fmt.Sprintf("%s.sql", n)
	path := filepath.Join(m.path, filename)

	content := "-- UP --\n\n-- DOWN --"

	saveFile(path, content)
}

func (m *Migrations) Migrate() error {
	if err := m.driver.CreateMigrationsTable(); err != nil {
		return err
	}

	migrations, err := m.GetUnexecutedMigrations()
	batch := m.nextBatch()

	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsToRun
	}

	for _, mg := range migrations{
		fmt.Println(mg.Name())
		q := mg.GetUpQuery()
		err := m.driver.Run(q)

		if err != nil {
			fmt.Printf("Failed migrating %s\n", mg.Name())
			log.Fatal(err)
			break
		}

		m.logMigration(mg, batch)
	}

	return nil
}

func (m *Migrations) Rollback() error {
	migrations, err := m.GetMigrationsReverse()

	if err != nil {
		return err
	}

	// Select only the migrations from the last batch
	mgs := m.GetMigrationsInBatch(m.currentBatch())

	for _, mg := range migrations {
		if !slices.Contains(mgs, mg.Name()) {
			continue
		}
		
		fmt.Println(mg.Name())
		q := mg.GetDownQuery()
		err := m.driver.Run(q)

		if err != nil {
			log.Fatal(err)
			break
		}

		m.removeMigration(mg)
	}

	return nil
}

// Drop all tables and migrate
func (m *Migrations) Fresh() {
	m.driver.Wipe()
	m.Migrate()
}

// Get migrations in order
func (m *Migrations) GetMigrations() ([]Migration, error) {
	files, err := os.ReadDir(m.path)

	if err != nil {
		return nil, err
	}

	migrations := []Migration{}

	for _, f := range files {
		if !f.IsDir() {
			path := filepath.Join(m.path, f.Name())
			migrations = append(migrations, Migration{Path: path})
		}
	}

	// Let them know we have no migrations
	if len(migrations) == 0 {
		return nil, ErrNoMigrations
	}

	return migrations, nil
}

// Get migrations in reverse order. Mostly for rollbacks
func (m *Migrations) GetMigrationsReverse() ([]Migration, error) {
	mg, err := m.GetMigrations()

	if err != nil {
		return nil, err
	}

	slices.Reverse(mg)

	return mg, nil
}

// Get the migrations that have not been run yet
func (m *Migrations) GetUnexecutedMigrations() ([]Migration, error) {
	mgs, err := m.GetMigrations()

	if err != nil {
		return nil, err
	}

	existing := m.GetExistingMigrations()

	un := []Migration{}

	for _, mg := range mgs {
		if !slices.Contains(existing, mg.Name()) {
			un = append(un, mg)
		}
	}

	return un, nil
}

func (m *Migrations) logMigration(mg Migration, batch int) {
	q := fmt.Sprintf("insert into migrations (migration, batch) values ('%s', %d)", mg.Name(), batch)
	m.driver.Run(q)
}

func (m *Migrations) removeMigration(mg Migration) {
	q := fmt.Sprintf("delete from migrations where migration = '%s'", mg.Name())
	m.driver.Run(q)
}

func (m *Migrations) currentBatch() int {
	db := m.driver.GetConnection()

	r := db.QueryRow("select max(batch) from migrations")

	var batch sql.NullInt64
	err := r.Scan(&batch)

	if err != nil {
		log.Fatal(err)
	}

	if !batch.Valid {
		return 0
	}

	return int(batch.Int64)
}

func (m *Migrations) nextBatch() int {
	return m.currentBatch() + 1
}

func (m *Migrations) GetExistingMigrations() []string {
	db := m.driver.GetConnection()

	r, err := db.Query("select migration from migrations")

	// If we don't get any rows return 1
	if err == sql.ErrNoRows {
		return []string{}
	}

	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()

	ms := []string{}

    for r.Next() {
        var n string
        if err := r.Scan(&n); err != nil {
            panic(err)
        }
        
		ms = append(ms, n)
    }

    if err := r.Err(); err != nil {
        log.Fatal(err)
    }

	return ms
}

func (m *Migrations) GetMigrationsInBatch(batch int) []string {
	db := m.driver.GetConnection()

	r, err := db.Query("select migration from migrations where batch = ?", batch)

	if err == sql.ErrNoRows {
		return []string{}
	}

	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()

	ms := []string{}

    for r.Next() {
        var n string
        if err := r.Scan(&n); err != nil {
            panic(err)
        }
        
		ms = append(ms, n)
    }

    if err := r.Err(); err != nil {
        log.Fatal(err)
    }

	return ms
}

func getMigrationFileName(name string) string {
	time := time.Now().Format("2006_01_02_150405")

	return fmt.Sprintf("%s_%s", time, name)
}