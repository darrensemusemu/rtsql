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
	return &postgresDB{
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
			fmt.Printf("received: %v\n", notication)
			break
		}
	}
}

func (p *postgresDB) AddTrigger(ctx context.Context, name string) error {
	return nil
}

func (p *postgresDB) RemoveTrigger(ctx context.Context) error {
	return nil
}

func (p *postgresDB) Ready(ctx context.Context) bool {
	return dbReady
}

func (p *postgresDB) OnUpdate() error {
	panic("postres onupdate func: not implemented")
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
