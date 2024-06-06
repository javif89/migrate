package migrate

import (
	"os"
	"path/filepath"
	"strings"
)

type Migration struct {
	Path string
}

func (m *Migration) Name() string {
	return strings.Trim(filepath.Base(m.Path), ".sql")
}

func (m *Migration) GetContent() (string, error) {
	content, err := os.ReadFile(m.Path)

	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (m *Migration) GetUpQuery() string {
	c, err := m.GetContent()

	if err != nil {
		return ""
	}

	parts := strings.Split(c, "-- DOWN --")

	q := strings.Replace(parts[0], "-- UP --\n", "", -1)
	q = strings.Trim(q, "\n")
	q = strings.TrimSpace(q)

	return q
}

func (m *Migration) GetDownQuery() string {
	c, err := m.GetContent()

	if err != nil {
		return ""
	}

	parts := strings.Split(c, "-- DOWN --")

	q := strings.Replace(parts[1], "-- DOWN --\n", "", -1)
	q = strings.Trim(q, "\n")
	q = strings.TrimSpace(q)

	return q
}