package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func RunMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(files)

	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		if len(body) == 0 {
			continue
		}

		if _, err := db.Exec(string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", path, err)
		}
	}

	return nil
}
