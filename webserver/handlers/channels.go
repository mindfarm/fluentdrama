package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	data "github.com/mindfarm/fluentdrama/webserver/repository/postgres"
)

type handlerData struct {
	ds *data.PGCustomerRepo
}

// NewChanData -
// ignore unexported linting error
// nolint:revive
func NewHandlerData(ds *data.PGCustomerRepo) *handlerData {
	return &handlerData{ds: ds}
}

// GetChannels -
func (hd *handlerData) GetChannels(w http.ResponseWriter, r *http.Request) {
	// return the list of channels that we have logs on
	// only GET allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Bad request - Go away!", http.StatusMethodNotAllowed)
		return
	}
	// Get channels from Datastore
	channels, err := hd.ds.GetChannels(context.Background())
	if err != nil {
		log.Printf("ERROR getting channels in GetChannels handler %v", err)
		return
	}

	resp, err := json.Marshal(struct {
		C []string `json:"channels"`
	}{channels})
	if err != nil {
		log.Printf("ERROR marshalling channels in GetChannels handler %v", err)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("ERROR writing channels in GetChannels handler %v", err)
		return
	}
}

// No query with a total length > maxQueryLength should be allowed
const maxQueryLength = 256

// Root route handler - default routes
func (hd *handlerData) Root(w http.ResponseWriter, r *http.Request) {
	path := []rune(r.URL.Path)

	if len(path) > maxQueryLength {
		return
	}

	// Is this a channel request
	if len(path) > 1 && (path[0] == '#' || (path[0] == '/' && path[1] == '#')) {
		// split on '/'
		chunks := []string{}
		holder := []rune{}
		for i := 1; i < len(path); i++ {
			if path[i] == '/' {
				chunks = append(chunks, string(holder))
				holder = []rune{}
				continue
			}
			holder = append(holder, path[i])
		}
		if len(holder) > 0 {
			chunks = append(chunks, string(holder))
		}
		// channel, date, nick, time will be after a ?
		if len(chunks) == 1 {
			// default to today
			chunks = append(chunks, strings.Split(time.Now().UTC().String(), " ")[0])
		}
		channel := chunks[0]
		// YYYY-MM-DD
		date, err := time.Parse("2006-01-02", chunks[1])
		if err != nil {
			log.Printf("ERROR parsing date: %v", err)
			http.Error(w, "Bad date supplied", http.StatusBadRequest)
			return
		}
		var nick string
		if len(chunks) > 2 {
			nick = chunks[2]
		}
		logs, err := hd.ds.GetChannelLogs(context.Background(), channel, nick, date)
		if err != nil {
			log.Printf("ERROR getting channel logs: %v", err)
			http.Error(w, "Bad channel or nick supplied", http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(struct {
			L []map[string]string `json:"logs"`
		}{logs})
		if err != nil {
			log.Printf("ERROR marshalling logs in GetChannelLogs handler %v", err)
			return
		}
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("ERROR writing logs in GetChannelLogs handler %v", err)
			return
		}

	}
}
