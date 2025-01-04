package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	charmLog "github.com/charmbracelet/log"
	"github.com/gorilla/mux"
	"github.com/japhy-tech/backend-test/database_actions"
	"github.com/japhy-tech/backend-test/internal"
	"./tests"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MysqlDSN   = "root:root@(mysql-test:3306)/?parseTime=true" // Notez l'absence de `core` pour la première connexion
	ApiPort    = "5000"
	BreedsFile = "./breeds.csv"
)

func main() {
	logger := charmLog.NewWithOptions(os.Stderr, charmLog.Options{
		Formatter:       charmLog.TextFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "🧑‍💻 backend-test",
		Level:           charmLog.DebugLevel,
	})

	db, err := sql.Open("mysql", MysqlDSN)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to MySQL: %s", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS core;")
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create database: %s", err.Error()))
		os.Exit(1)
	}
	logger.Info("Database `core` ensured to exist")

	MysqlDSNWithDB := "root:root@(mysql-test:3306)/core?parseTime=true"
	db, err = sql.Open("mysql", MysqlDSNWithDB)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to `core` database: %s", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to ping database: %s", err.Error()))
		os.Exit(1)
	}
	logger.Info("Connected to database `core`")

	err = database_actions.InitMigrator(MysqlDSNWithDB)
	if err != nil {
		logger.Fatal(err.Error())
	}

	msg, err := database_actions.RunMigrate("up", 0)
	if err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(msg)
	}

	err = database_actions.ImportBreeds(db, BreedsFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to import breeds: %s", err.Error()))
		os.Exit(1)
	}
	logger.Info("Breeds imported successfully")

	app := internal.NewApp(logger)
	app.DB = db

	r := mux.NewRouter()
	app.RegisterRoutes(r.PathPrefix("/v1").Subrouter())

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	err = http.ListenAndServe(
		net.JoinHostPort("", ApiPort),
		r,
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to start server: %s", err.Error()))
	}

	logger.Info(fmt.Sprintf("Service started and listen on port %s", ApiPort))
	logger.Info(fmt.Sprintf("Start testing"))
	tests.RunTests()
}
