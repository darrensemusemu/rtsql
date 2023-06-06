package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	minReconnectInterval = 10 * time.Second
	maxReconnectInterval = time.Minute
)

var dbReady = false // FIXME: probably not safe for use by mutiple go routines

type postgresDB struct {
	db         *sqlx.DB
	pqListener *pq.Listener
}

var _ Repository = (*postgresDB)(nil)

func NewPostgresDB(connString string) *postgresDB {
	pqListener := pq.NewListener(connString, minReconnectInterval, maxReconnectInterval, pqListerCallback)

	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		panic(err)
	}

	return &postgresDB{
		db:         db,
		pqListener: pqListener,
	}
}

func (p *postgresDB) Listen(ctx context.Context, channel string) error {
	err := p.pqListener.Listen(channel) // TODO: might reurn pq.Error please test
	if err != nil {
		return err
	}

	for {
		select {
		case notication := <-p.pqListener.Notify:
			fmt.Printf("received: %+v\n", notication)
			break
		}
	}
}

// TODO: move to db on ready
const postgresTriggerFn = `
CREATE OR REPLACE FUNCTION rtsql_event_trigger_fn()
RETURNS TRIGGER AS $$
DECLARE
	channel text := TG_ARGV[0];
BEGIN
	PERFORM pg_notify(channel, row_to_json(NEW)::text);
	return NULL;
END;
$$ LANGUAGE plpgsql;
`

func (p *postgresDB) AddTrigger(ctx context.Context, cfg ConfigModelTable) error {
	// FIXME: please lets fix this (possible sql injection ???)
	sql := fmt.Sprintf(`CREATE OR REPLACE TRIGGER rtsql_event_notify_%s_tbl
	AFTER INSERT
	ON "%s"
	FOR EACH ROW
	EXECUTE PROCEDURE rtsql_event_trigger_fn('rtsql_event_channel');
	`, cfg.Name, cfg.Name)

	_, err := p.db.ExecContext(ctx, postgresTriggerFn+sql)
	return err
}

func (p *postgresDB) RemoveTrigger(ctx context.Context) error {
	return nil
}

func (p *postgresDB) Ready(ctx context.Context) bool {
	return dbReady && p.db.Ping() == nil
}

func (p *postgresDB) OnUpdate() error {
	panic("postres onupdate func: not implemented")
}

func (p *postgresDB) Close() error {
	err := p.pqListener.Close()
	return err
}

func pqListerCallback(event pq.ListenerEventType, err error) {
	switch event {
	case pq.ListenerEventConnected, pq.ListenerEventReconnected:
		dbReady = true
		break
	case pq.ListenerEventDisconnected, pq.ListenerEventConnectionAttemptFailed:
		dbReady = false
		break
	default:
		fmt.Printf("pq listener callback: unknown event (%v)", event)
		os.Exit(1)
	}
}
