package database_actions

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var driver database.Driver

// InitMigrator initiates values essential for migrations
func InitMigrator(dsnMigrate string) error {
	var err error
	db, err := sql.Open("mysql", dsnMigrate)
	if err != nil {
		return fmt.Errorf("error while opening db connection: %w", err)
	}
	driver, err = mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("error while instanciating migration driver: %w", err)
	}

	return nil
}

// RunMigrate performs all or only some up/down migrations
//
// Default 'steps' as 0 (runs all migrations)
func RunMigrate(migrationType string, steps int) (string, error) {
	m, err := migrate.NewWithDatabaseInstance(
		"file://database_actions/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return "", fmt.Errorf("error while instanciating new migration ("+migrationType+") with DB : %w", err)
	}

	if steps != 0 {
		m.Steps(steps)
	} else {
		if migrationType == "up" {
			err = m.Up()
			if errors.Is(err, migrate.ErrNoChange) {
				return "Migration(s) : " + migrate.ErrNoChange.Error(), nil
			}
			if err != nil {
				return "", fmt.Errorf("error while running up migration(s): %w", err)
			}
		} else if migrationType == "down" {
			err = m.Down()
			if errors.Is(err, migrate.ErrNoChange) {
				return "Migration(s) : " + migrate.ErrNoChange.Error(), nil
			}
			if err != nil {
				return "", fmt.Errorf("error while running down migration(s): %w", err)
			}
		} else {
			return "", fmt.Errorf("error unknown migration type: " + migrationType)
		}
	}

	return migrationsSuccessMessage(migrationType, steps), nil
}

func migrationsSuccessMessage(migrationType string, steps int) string {
	msg := "Successfully ran"
	if steps == 0 {
		return msg + " all " + migrationType + " migrations"
	}
	if steps == 1 {
		return msg + " 1 " + migrationType + " migration"
	}

	return msg + " " + strings.Trim(strconv.Itoa(steps), "-") + " " + migrationType + " migrations"
}
