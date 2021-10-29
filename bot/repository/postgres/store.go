package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type pgCustomerRepo struct {
	dbHandler *pgxpool.Pool
}

// NewPgCustomerRepo -
// Ignore unexpected type linter issue
// nolint:revive
func NewPgCustomerRepo(connString string) (*pgCustomerRepo, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err

	}
	/*
		config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			conn.ConnInfo().RegisterDataType(pgtype.DataType{
				Value: &MyTid{&pgtype.TID{}},

				Name: "tid",
				OID:  pgtype.TIDOID,
			})
			return nil

		}
	*/

	db, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err

	}

	return &pgCustomerRepo{
		dbHandler: db,
	}, nil

}

// AddChannel -
func (p *pgCustomerRepo) AddChannel(ctx context.Context, channel string) error {
	_, err := p.dbHandler.Query(ctx, `INSERT INTO channels(name) VALUES($1)`, channel)
	if err != nil {
		return fmt.Errorf("adding channel %q produced %w", channel, err)
	}
	return err
}

// GetChannels -
func (p *pgCustomerRepo) GetChannels(ctx context.Context) ([]string, error) {
	rows, err := p.dbHandler.Query(ctx, `SELECT name from channels`)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch channels with error %w`, err)
	}

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
func (p *pgCustomerRepo) AddLog(ctx context.Context, channel, username, said string) error {
	_, err := p.dbHandler.Query(ctx, `INSERT INTO logs(channel, username,  said) VALUES($1, $2, $3)`, channel, username, said)
	if err != nil {
		return fmt.Errorf("adding log %q %q %q produced %w", channel, username, said, err)
	}
	return err
}

// GetChannelLogsByTime -
func (p *pgCustomerRepo) GetChannelLogsByTime(ctx context.Context, channel string, start, finish time.Time) ([]map[string]string, error) {
	rows, err := p.dbHandler.Query(ctx, `SELECT  nick, stamp, said FROM channels WHERE channel=$1 stamp BETWEEN $2 AND $3`, channel, start, finish)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch channels with error %w`, err)
	}

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
