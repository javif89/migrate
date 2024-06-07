package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/javif89/migrate/database"
)

func TestCreateMigration(t *testing.T) {
	d := t.TempDir()
	m := New(d, database.DriverTesting, database.Config{})

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