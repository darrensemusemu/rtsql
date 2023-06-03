package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
)

type config struct {
	dbType     dbType
	dbConn     string
	configPath string // TODO: add support for json, toml, ...
}

func main() {
	cfg := config{dbType: postresDb}

	flag.Var(&cfg.dbType, "type", "db type 'postgres', 'sqlite', 'sqlserver' (default \"postgres\")")
	flag.StringVar(&cfg.dbConn, "dsn", "postgres://postgres:root@localhost:5432/?sslmode=disable", "db connection string")
	flag.StringVar(&cfg.configPath, "config", "config.yaml", "rtsql config file")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Printf("exited with err: %s", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	var repo Repository
	var err error
	var db *sqlx.DB

	// connect db
	switch cfg.dbType {
	case postresDb:
		db, err = sqlx.Open("pgx", "dsn")
		repo = NewPostgresDB(db)
	default:
		return fmt.Errorf("unsupported db: %s\n", cfg.dbType)
	}
	if err != nil {
		return err
	}

	// load config file
	configContents, err := os.ReadFile(cfg.configPath)
	if err != nil {
		return err
	}
	var model ConfigModel
	if err := yaml.Unmarshal(configContents, &model); err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		x := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "lfg %v", repo.Ready())
		})
		http.ListenAndServe(":8080", x)
	}()

	err = <-errChan
	return err
}

type ConfigModel struct {
	Migrations string             `yaml:"migrations"`
	Tables     []ConfigModelTable `yaml:"tables"`
}

type ConfigModelTable struct {
	Name    string              `yaml:"name"`
	On      []string            `yaml:"on"` // FIXME: way to type check this
	Actions []ConfigModelAction `yaml:"actions"`
}

type ConfigModelAction struct {
	To          string `yaml:"to"`
	ContentType string `yaml:"content_type"`
}
