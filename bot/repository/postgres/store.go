package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	//"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq" //nolint:revive
)

type pgCustomerRepo struct {
	//dbHandler *pgxpool.Pool
	dbHandler *sql.DB
}

// NewPgCustomerRepo -
// Ignore unexpected type linter issue
// nolint:revive
func NewPgCustomerRepo(connString string) (*pgCustomerRepo, error) {
	//config, err := pgxpool.ParseConfig(connString)
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err

	}

	/*
		db, err := pgxpool.ConnectConfig(context.Background(), config)
		if err != nil {
			return nil, err
		}
	*/

	return &pgCustomerRepo{
		dbHandler: conn,
	}, nil

}

// AddChannel -
func (p *pgCustomerRepo) AddChannel(ctx context.Context, channel string) error {
	_, err := p.dbHandler.Query(`INSERT INTO channels(name) VALUES($1)`, channel)
	if err != nil {
		return fmt.Errorf("adding channel %q produced %w", channel, err)
	}
	return err
}

// GetChannels -
func (p *pgCustomerRepo) GetChannels(ctx context.Context) ([]string, error) {
	rows, err := p.dbHandler.Query(`SELECT name from channels`)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch channels with error %w`, err)
	}
	defer rows.Close()

	channels := []string{}
	var channel sql.NullString
	for rows.Next() {
		err := rows.Scan(&channel)
		if err != nil {
			log.Printf("Unable to scan channel with error %v", err)
			continue
		}
		channels = append(channels, channel.String)
	}
	return channels, nil
}

// AddLog -
func (p *pgCustomerRepo) AddLog(ctx context.Context, channel, nick, said string) error {
	rows, err := p.dbHandler.Query(`INSERT INTO logs(channel, nick,  said) VALUES($1, $2, $3)`, channel, nick, said)
	if err != nil {
		return fmt.Errorf("adding log %q %q %q produced %w", channel, nick, said, err)
	}
	defer rows.Close()
	return err
}

// GetChannelLogsByTime -
func (p *pgCustomerRepo) GetChannelLogsByTime(ctx context.Context, channel string, start, finish time.Time) ([]map[string]string, error) {
	rows, err := p.dbHandler.Query(`SELECT  nick, stamp, said FROM channels WHERE channel=$1 stamp BETWEEN $2 AND $3`, channel, start, finish)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch channels with error %w`, err)
	}
	defer rows.Close()

	logs := []map[string]string{}
	var nick sql.NullString
	var said sql.NullString
	var stamp sql.NullTime
	for rows.Next() {
		err := rows.Scan(&nick, &stamp, &said)
		if err != nil {
			log.Printf("Unable to scan channel with error %v", err)
			continue
		}
		logs = append(logs, map[string]string{"time": stamp.Time.String(), "nick": nick.String, "said": said.String})
	}
	return logs, nil
}
