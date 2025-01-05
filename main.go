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
	"github.com/Asto-42/TechTestJaphy/database_actions"
	"github.com/Asto-42/TechTestJaphy/internal"
	"github.com/Asto-42/TechTestJaphy/tests"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MysqlDSN   = "root:root@(mysql-test:3306)/?parseTime=true"
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

	go func() {
		logger.Info(fmt.Sprintf("API port: %s", ApiPort))
		logger.Info(fmt.Sprintf("Server is listening on http://127.0.0.1:%s", ApiPort))
		err := http.ListenAndServe(
			net.JoinHostPort("127.0.0.1", ApiPort),
			r,
		)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to start server: %s", err.Error()))
		}
	}()

	go func() {
		logger.Info("Waiting for server readiness...")
		for i := 1; i <= 10; i++ {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", ApiPort))
			if err == nil && resp.StatusCode == http.StatusOK {
				logger.Info("Server is ready. Starting tests.")
				tests.StartTests()
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {}
}
