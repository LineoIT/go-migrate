package migrate

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	TIME_LAYOUT = "20060201150405"
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

func Create(dir, name string) {
	baseName := fmt.Sprintf("%s/%s_%s", dir, time.Now().Format(TIME_LAYOUT), name)
	file, err := os.Create(fmt.Sprintf("%s.up.sql", baseName))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer file.Close()
	_, err = os.Create(fmt.Sprintf("%s.down.sql", baseName))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	fmt.Printf(" migration %s.down.sql\n", baseName)
}
