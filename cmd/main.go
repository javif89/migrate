package main

import (
	"fmt"
	"log"
	"os"

	"github.com/javif89/dotenv"
	"github.com/javif89/migrate"
	"github.com/javif89/migrate/database"
	"github.com/urfave/cli/v2"
)

func main() {
    var m migrate.Migrations

	if envExists() {
        f := dotenv.Load(".env")

        m = *migrate.New(f.Get("MIGRATIONS_PATH"), database.DriverName(f.Get("DB_DRIVER")), database.Config{
            Host:     f.Get("DB_HOST"),
            Port:     f.Get("DB_PORT"),
            Username: f.Get("DB_USERNAME"),
            Password: f.Get("DB_PASSWORD"),
            Database: f.Get("DB_DATABASE"),
        })
    }

	app := &cli.App{
		Name:  "migrate",
		Usage: "Create and run database migrations for your app",
		Action: func(ctx *cli.Context) error {
			if !envExists() {
				fmt.Println(".env file not found. Please run migrate init")
				return nil
			}

			fmt.Println("Running migrations")

			err := m.Migrate()

			if err == migrate.ErrNoMigrationsToRun {
				fmt.Println("No migrations to run")
				return nil
			}

			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize a .env file with the variables you need to set up migrations",
				Action: func(cCtx *cli.Context) error {
					if envExists() {
						fmt.Println(".env file already exists")
						return nil
					}

					initEnv()

					return nil
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new migration",
				Action: func(cCtx *cli.Context) error {
					name := cCtx.Args().First()
					m.CreateMigration(name)
					return nil
				},
			},
			{
				Name:    "rollback",
				Aliases: []string{"r"},
				Usage:   "Rollback last migration batch",
				Action: func(cCtx *cli.Context) error {
					if !envExists() {
						fmt.Println(".env file not found. Please run migrate init")
						return nil
					}

					fmt.Println("Rolling back")

					err := m.Rollback()

					if err != nil {
						log.Fatal(err)
					}

					return nil
				},
			},
			{
				Name:    "fresh",
				Aliases: []string{"f"},
				Usage:   "Delete all tables and migrate again",
				Action: func(cCtx *cli.Context) error {
					if !envExists() {
						fmt.Println(".env file not found. Please run migrate init")
						return nil
					}

					fmt.Println("Deleting all tables")

					m.Fresh()

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func envExists() bool {
	_, err := os.Stat(".env")

	return !os.IsNotExist(err)
}

func initEnv() {
	os.Create(".env")
	f := dotenv.Load(".env")

	f.Add("DB_HOST", "localhost")
	f.Add("DB_PORT", "3306")
	f.Add("DB_USERNAME", "user")
	f.Add("DB_PASSWORD", "pass")
	f.Add("DB_DATABASE", "mydb")
	f.Add("DB_DRIVER", "mysql")
	f.Add("MIGRATIONS_PATH", "./database/migrations")

	f.Save()
}
