package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/LineoIT/go-migrate"
	"log"
	"os"
	_ "github.com/lib/pq"
	"strings"
)

const HELP = `
Usage:
	console <cli> [arguments]
The commands are:
	help	Help
	up	Migration database
		--dsn 			<string> Database source url
		--dir 			<string> Migration files folder
		--production 	<string> Specify production environment, default true
	down	Migration rollback database
		--dsn 			<string> Database source url
		--dir 			<string> Migration files folder
		--production 	<string> Specify production environment, default true
	create  CreateFile user
		--name	<string> file name
		--dir	<string> Migration files folder
`
const (
	ACTION_HELP = "help"
	ACTION_UP = "up"
	ACTION_DOWN = "down"
	ACTION_CREATE = "create"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Print(HELP)
		os.Exit(0)
	}

	switch os.Args[1] {
	case ACTION_HELP:
		fmt.Print(HELP)
	case ACTION_UP:
		migrateDatabaseCmd(ACTION_UP)
	case ACTION_DOWN:
		migrateDatabaseCmd(ACTION_DOWN)
	case ACTION_CREATE:
		createMigrationCmd()
	default:
		fmt.Println("Command not found")
		fmt.Print(HELP)
	}
}

func migrateDatabaseCmd(action string){
	cmd := flag.NewFlagSet(action, flag.ExitOnError)
	dsn := cmd.String("dsn", "", "database url")
	dir := cmd.String("dir", "", "migration files folder")
	isProd := cmd.Bool("production", true, "Production environment")
	msg := "Do you sure to migrate database in production mode?"
	if action == ACTION_DOWN {
		msg = "Do you really want to delete all tables in production mode?"
	}
	if confirm(msg,  *isProd) {
		if err := cmd.Parse(os.Args[2:]); err != nil {
			fmt.Println(err.Error())
			cmd.PrintDefaults()
			os.Exit(1)
		}

		if *dsn == ""  {
			fmt.Println("Error: database source required")
			os.Exit(1)
		}

		if *dir == ""  {
			fmt.Println("Error: migration files folder required")
			os.Exit(1)
		}

		isRollback := action == ACTION_DOWN

		if err := migrateTables(*dsn, *dir, isRollback); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}

}

func createMigrationCmd() {
	cmd := flag.NewFlagSet(ACTION_CREATE, flag.ExitOnError)
	name := cmd.String("name", "", "Migration name")
    dir := cmd.String("dir", "", "migration files folder")
	if err := cmd.Parse(os.Args[2:]); err != nil {
		fmt.Println(err.Error())
	}
	if *name == "" {
		fmt.Println("Migration name needed")
		cmd.PrintDefaults()
		os.Exit(1)
		return
	}
	if *dir == "" {
		fmt.Println("Migration output directory needed")
		cmd.PrintDefaults()
		os.Exit(1)
		return
	}
	migrate.CreateFile(*dir, *name)
}

func  migrateTables(dns, dir string, isRollback bool) error {
	migration, err := migrate.New("postgres", dns, dir)
	if err != nil {
		return err
	}
	defer migration.Close()
	if isRollback {
		return migration.Rollback()
	}
	return migration.Migrate()
}

func confirm(msg string, isProd bool) bool {
	if isProd {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Printf("%s [y/n]: ", msg)
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			response = strings.ToLower(strings.TrimSpace(response))

			if response == "y" || response == "yes" {
				return true
			} else if response == "n" || response == "no" {
				return false
			}
		}
	}
	return true
}