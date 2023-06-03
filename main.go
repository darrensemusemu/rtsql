package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	dbType dbType
	dbConn string
}

func main() {
	cfg := config{dbType: postresDb}

	flag.Var(&cfg.dbType, "type", "db type 'postgres', 'sqlite', 'sqlserver' (default \"postgres\")")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Printf("exited with err: %s", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	print("running")
	return nil
}
