package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	dbType     dbType
	dbConn     string
	configPath string // TODO: add support for json, toml, ...
}

const (
	rtsql_channel = "rtsql_user"
)

func main() {
	cfg := config{dbType: postresDb}

	flag.Var(&cfg.dbType, "type", "db type 'postgres', 'sqlite', 'sqlserver' (default \"postgres\")")
	flag.StringVar(&cfg.dbConn, "db-conn", "postgres://postgres:root@127.0.0.1:5432/rsql-test?sslmode=disable", "db connection string")
	flag.StringVar(&cfg.configPath, "config", "config.yaml", "rtsql config file")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Printf("exited with err: %s", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	ctx := context.Background()

	var repo Repository
	var err error

	// connect db
	switch cfg.dbType {
	case postresDb:
		repo = NewPostgresDB(cfg.dbConn)
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

	errCh := make(chan error)

	go func() {
		if listenErr := repo.Listen(ctx, rtsql_channel); listenErr != nil {
			errCh <- listenErr
			close(errCh)
		}
	}()

	go func() {
		x := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "DB ready:  %v\n", repo.Ready(ctx))
		})
		fmt.Printf("Listening on port 8080\n")
		http.ListenAndServe(":8080", x)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		break
	}

	return err
}

func parseConfigModel(model ConfigModel) error {
	if model.Migrations == "" {
		return errors.New("parse config-model error: migrations table not specified")
	}

	// FIXME: handle other dbs
	for _, m := range model.Tables {
		_ = m.Name
	}

	return nil
}

type ConfigModel struct {
	Migrations string             `yaml:"migrations"` // TODO: handle migrations
	Tables     []ConfigModelTable `yaml:"tables"`
}

type ConfigModelTable struct {
	Name    string              `yaml:"name"`
	On      []dbOnType          `yaml:"on"`
	Actions []ConfigModelAction `yaml:"actions"`
}

type ConfigModelAction struct {
	To          string `yaml:"to"`
	ContentType string `yaml:"content_type"`
}
