// Package data - this user will only be granted READONLY access to the
// database. The database is written to by a different application and that
// application owns the schema.
package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" //nolint:revive
)

// PGCustomerRepo -
type PGCustomerRepo struct {
	DbHandler *sql.DB
}

// NewPGCustomerRepo -
// Ignore unexpected type linter issue
// nolint:revive
func NewPGCustomerRepo(connString string) (*PGCustomerRepo, error) {
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	return &PGCustomerRepo{
		DbHandler: conn,
	}, nil
}

// GetChannels -
func (p *PGCustomerRepo) GetChannels(ctx context.Context) ([]string, error) {
	rows, err := p.DbHandler.Query(`SELECT name FROM channels ORDER BY name ASC`)
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

// GetChannelLogs -
func (p *PGCustomerRepo) GetChannelLogs(ctx context.Context, channel, nick string, date time.Time) ([]map[string]string, error) {
	// channel is mandatory
	// nick is optional
	if channel == "" {
		return nil, fmt.Errorf("channel is mandatory")
	}

	finish := date.Add(24 * time.Hour)
	// ICK TWO DB lookups to find the boundaries, at least one of these should
	// be from a cache
	// TODO cache the start for each channel for faster lookup
	// get the time of the latest entry in the logs for this channel
	l, err := p.getBoundary(nick, channel, "last")
	if err != nil {
		return nil, err
	}
	// get the time of the earliest entry in the logs for this channel
	f, err := p.getBoundary(nick, channel, "first")
	if err != nil {
		return nil, err
	}
	if l.Sub(finish) < 0 {
		finish = l
	}

	// get the 24 hours before the end datetime
	start := finish.Add(-24 * time.Hour)
	if f.Sub(start) > 0 {
		start = f
		finish = start.Add(24 * time.Hour)
	}

	var rows *sql.Rows
	if nick == "" {
		rows, err = p.DbHandler.Query(`SELECT  nick, stamp, said FROM logs WHERE channel=$1 AND stamp BETWEEN $2 AND $3 ORDER BY stamp ASC`, channel, start, finish)
	} else {
		// only get the logs for the specified nick
		rows, err = p.DbHandler.Query(`SELECT  nick, stamp, said FROM logs WHERE channel=$1 AND nick=$2 AND stamp BETWEEN $3 AND $4 ORDER BY stamp ASC`, channel, nick, start, finish)
	}
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch channels with error %w`, err)
	}

	logs := []map[string]string{}
	var rnick sql.NullString
	var rsaid sql.NullString
	var rstamp sql.NullTime
	for rows.Next() {
		err := rows.Scan(&rnick, &rstamp, &rsaid)
		if err != nil {
			log.Printf("Unable to scan channel with error %v", err)
			continue
		}
		logs = append(logs, map[string]string{"Time": rstamp.Time.String(), "Nick": rnick.String, "Said": rsaid.String})
	}
	return logs, nil
}

func (p *PGCustomerRepo) getBoundary(nick, channel, order string) (time.Time, error) {
	direction := "DESC"
	if order == "first" {
		direction = "ASC"
	}
	var rows *sql.Rows
	var err error
	if nick != "" {
		query := fmt.Sprintf("SELECT stamp FROM logs WHERE channel=$1 AND nick=$2 ORDER BY stamp %s LIMIT 1", direction)
		rows, err = p.DbHandler.Query(query, channel, nick)
	} else {
		query := fmt.Sprintf("SELECT stamp FROM logs WHERE channel=$1  ORDER BY stamp %s LIMIT 1", direction)
		rows, err = p.DbHandler.Query(query, channel)
	}
	defer rows.Close()
	if err != nil {
		return time.Time{}, fmt.Errorf(`unable to fetch final time stamp in logs with error %w`, err)
	}
	var f sql.NullTime
	for rows.Next() {
		if err = rows.Scan(&f); err != nil {
			return time.Time{}, fmt.Errorf("error scanning final time in logs %w", err)
		}
	}
	if !f.Valid {
		// no channel
		return time.Time{}, fmt.Errorf("channel %s or nick %s does not exist", channel, nick)
	}
	return f.Time, nil
}
