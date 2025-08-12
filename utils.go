package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Version struct {
	Version int64
	IsDirty *bool
	Date    *time.Time
}

func (v *Version) Save(m *Migration) error {
	_, err := m.Exec(fmt.Sprintf(`
	update %s set dirty=coalesce($2, dirty) where version=$1`, m.Config.Table),
		v.Version,
		v.IsDirty,
	)
	return err
}

func (v *Version) Create(m *Migration) error {
	if err := m.QueryRow(
		fmt.Sprintf(`insert into %s(version)
		values($1) returning created_at;`, m.Config.Table),
		v.Version,
	).Scan(&v.Date); err != nil {
		return err
	}
	return nil
}

func CreateFile(dir, name string) error {
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create migrations directory: %w", err)
    }

    files, err := os.ReadDir(dir)
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %w", err)
    }

     nextVersion := 1
    for _, f := range files {
        parts := strings.SplitN(f.Name(), "__", 2)
        if len(parts) < 2 {
            continue
        }

        currentVersion, err := strconv.Atoi(parts[0])
        if err != nil {
            continue
        }
        if currentVersion >= nextVersion {
            nextVersion = currentVersion + 1
        }
    }


    // Sanitize migration name
    safeName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
    baseName := fmt.Sprintf("%03d__%s", nextVersion, safeName)

    // Create .up.sql
    upPath := filepath.Join(dir, fmt.Sprintf("%s.up.sql", baseName))
    if _, err := os.Stat(upPath); err == nil {
        return fmt.Errorf("up migration already exists: %s", upPath)
    }
    if err := os.WriteFile(upPath, []byte("-- Write your UP migration here\n"), 0644); err != nil {
        return fmt.Errorf("failed to create up migration file: %w", err)
    }

    // Create .down.sql
    downPath := filepath.Join(dir, fmt.Sprintf("%s.down.sql", baseName))
    if _, err := os.Stat(downPath); err == nil {
        return fmt.Errorf("down migration already exists: %s", downPath)
    }
    if err := os.WriteFile(downPath, []byte("-- Write your DOWN migration here\n"), 0644); err != nil {
        return fmt.Errorf("failed to create down migration file: %w", err)
    }

    fmt.Printf("Migration created:\n  %s\n  %s\n", upPath, downPath)
    return nil
}


