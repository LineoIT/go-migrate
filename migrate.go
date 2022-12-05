package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"
)

var ErrNoChange = errors.New("no change")
var ErrDirty = errors.New("migration is dirty")

type Migration struct {
	*sql.DB
	Config  *Config
	baseDir string
	driver  string
}

type Config struct {
	Table string
}

// New creates new migration
func New(driver, dsn, baseDir string) (*Migration, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	m := &Migration{DB: db, Config: &Config{
		Table: "migrations",
	}, baseDir: baseDir, driver: driver}
	return m, m.createTable()
}

// * Run migrations
func (m *Migration) Migrate() error {
	files, lastVersion, err := m.getFiles("up")
	if err != nil {
		return err
	}
	var newVersion int
	for i, f := range files {
		b, err := os.ReadFile(m.baseDir + "/" + f.Name())
		if err != nil {
			return err
		}
		newVersion, err = strconv.Atoi(strings.Split(f.Name(), "_")[0])
		lastVersion.Version = int64(newVersion)
		if err != nil {
			return err
		}
		_, err = m.Exec(string(b))
		if err != nil {
			d := true
			lastVersion.IsDirty = &d
			lastVersion.Create(m)
			return fmt.Errorf(string(b)+"%s\n", err)
		}
		lastVersion.Create(m)
		fmt.Printf("%d/%d  %s\n", i+1, len(files), f.Name())
	}
	return nil
}

// ! Rollback all migrations from database
func (m *Migration) Rollback() error {
	files, lastVersion, err := m.getFiles("down")
	if err != nil {
		return err
	}
	if lastVersion.Version == 0 || len(files) == 0 {
		fmt.Println(ErrNoChange.Error())
		return nil
	}
	for i, f := range files {
		b, err := os.ReadFile(m.baseDir + "/" + f.Name())
		if err != nil {
			return err
		}
		_, err = m.Exec(string(b))
		if err != nil {
			return fmt.Errorf(string(b)+"%s\n", err)
		}
		fmt.Printf("%d/%d  %s\n", i+1, len(files), f.Name())
	}
	err = m.Clean()
	return err
}

// Clean delete all migrations versions
func (m *Migration) Clean() error {
	q := fmt.Sprintf(`truncate table %s;`, m.Config.Table)
	if m.driver == "sqlite3" || m.driver == "sqlite" {
		q = fmt.Sprintf(`delete from %s;`, m.Config.Table)
	}
	_, err := m.Exec(q)
	if err != nil {
		fmt.Println(q)
		return err
	}
	return nil
}

// GetLastVersion get last migration version
func (m *Migration) GetLastVersion() (Version, error) {
	var v Version
	if err := m.QueryRow(fmt.Sprintf("select * from %s order by created_at desc limit 1", m.Config.Table)).
		Scan(&v.Version, &v.IsDirty, &v.Date); err != nil {
		if err != sql.ErrNoRows {
			return Version{}, errors.New("migration " + err.Error())
		}
	}
	if v.IsDirty != nil {
		if *v.IsDirty {
			return Version{}, ErrDirty
		}
	}
	return v, nil
}

// GetVersions get all migrations version
func (m *Migration) GetVersions() ([]Version, error) {
	rows, err := m.Query(fmt.Sprintf("select * from %s", m.Config.Table))
	if err != nil {
		return []Version{}, err
	}
	result := make([]Version, 0)
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.Version, &v.IsDirty, &v.Date); err != nil {
			if err != sql.ErrNoRows {
				return []Version{}, errors.New("migration " + err.Error())
			}
		}
		result = append(result, v)
	}
	return result, nil
}

func (m *Migration) createTable() error {
	_, err := m.Exec(fmt.Sprintf(`create table if not exists %s(
		 version varchar(60) not null unique,
		 dirty bool default(false),
		 created_at timestamp default now()
	     );
		`, m.Config.Table))
	return err
}

func (m *Migration) getFiles(migrateType string) (fss []fs.DirEntry, lastVersion Version, err error) {
	files, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fss, lastVersion, err
	}
	lastVersion, err = m.GetLastVersion()
	if err != nil {
		return fss, lastVersion, err
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), "."+migrateType+".sql") {
			if migrateType == "down" {
				fss = append(fss, f)
			} else {
				currentVersion, _ := strconv.Atoi(strings.Split(f.Name(), "_")[0])
				if lastVersion.Version == 0 || int64(currentVersion) > lastVersion.Version {
					fss = append(fss, f)
				}
			}
		}
	}
	if migrateType == "down" {
		sort.Slice(fss, func(i, j int) bool {
			return fss[i].Name() > fss[j].Name()
		})
	}
	return fss, lastVersion, nil
}
