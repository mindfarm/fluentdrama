package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

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
		S []string `json:"channels"`
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
