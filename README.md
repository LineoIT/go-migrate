# Golang migrate

### Install

```
go get -u github.com/LineoIT/go-migrate
```

### How to use

```go
package main

import (
	"github.com/LineoIT/go-migrate"
	"log"
	_ "github.com/lib/pq"
)

func main() {
	m, err := migrate.New("postgres",
		"postgres://postgres:postgres@localhost/demo?sslmode=disable",
		"./migrations",
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Migrate(); err != nil {
		panic(err)
	}
}

```

**Note**: Do not forget import db driver `github.com/lib/pq` for postgres
