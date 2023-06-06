package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"gopkg.in/yaml.v3"
)

type config struct {
	dbType     dbType
	dbConn     string
	configPath string // TODO: add support for json, toml, ...
}

const (
	rtsql_channel = "rtsql_event_channel"
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
	errCh := make(chan error)

	// setup db
	repo, err := setupRepo(cfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	//
	go func() {
		if listenErr := repo.Listen(ctx, rtsql_channel); listenErr != nil {
			errCh <- listenErr
			close(errCh)
		}
	}()

	go func() {
		if errConf := runConfig(ctx, cfg.configPath, repo); errConf != nil {
			errCh <- errConf
			close(errCh)
		}
	}()

	go func() {
		fmt.Printf("Listening on port 8080\n")
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "DB ready:  %v\n", repo.Ready(ctx))
		}))
	}()

	// err checking goroutines
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		break
	}

	return err
}

func setupRepo(cfg config) (repo Repository, err error) {
	switch cfg.dbType {
	case postresDb:
		repo = NewPostgresDB(cfg.dbConn)
	default:
		return repo, fmt.Errorf("unsupported db: %s\n", cfg.dbType)
	}

	return repo, err
}

func runConfig(ctx context.Context, path string, repo Repository) error {
	configContents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var model ConfigModel
	if err := yaml.Unmarshal(configContents, &model); err != nil {
		return err
	}

	backoff := time.Second // TODO: implment a time out flag
	for {
		if repo.Ready(ctx) {
			break
		}
		time.Sleep(backoff)
		backoff = backoff << 1
	}

	for _, m := range model.Tables {
		err := repo.AddTrigger(ctx, m) // TODO: support other actions e.g. on update
		if err != nil {
			return err
		}
	}

	return nil
}

type ConfigModel struct {
	Migrations string             `yaml:"migrations"` // TODO: handle migrations & validate fields
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
