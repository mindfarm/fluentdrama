// Package data - this user will only be granted READONLY access to the
// database. The database is written to by a different application and that
// application owns the schema.
package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" //nolint:revive
)

// PGCustomerRepo -
type PGCustomerRepo struct {
	DbHandler *sql.DB
}

// NewPgCustomerRepo -
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
	rows, err := p.DbHandler.Query(`SELECT name from channels`)
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
