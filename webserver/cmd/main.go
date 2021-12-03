package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mindfarm/fluentdrama/webserver/handlers"
	data "github.com/mindfarm/fluentdrama/webserver/repository/postgres"
)

func main() {
	dbURI, ok := os.LookupEnv("DBURI")
	if !ok {
		log.Fatalf("DBURI is not set")
	}

	// Datastore
	ds, err := data.NewPGCustomerRepo(dbURI)
	if err != nil {
		log.Fatalf("Unable to connect to datastore with error %v", err)
	}

	// Get port from env
	var rPort string
	if rPort, ok = os.LookupEnv("HTTP_PORT"); !ok {
		log.Fatal("HTTP_PORT required for HTTP server to listen on")
	}

	// Port validation - has to be converted to an int for the checks
	var port int
	if port, err = strconv.Atoi(rPort); err != nil {
		log.Fatal("HTTP_PORT must be an integer")
	}
	// PORT must be non-privileged and legit
	if port <= 1024 || port >= 65535 {
		log.Fatal("HTTP_PORT must be between 1024 and 65535 (exclusive)")
	}

	mux := http.NewServeMux()

	c := handlers.NewHandlerData(ds)
	mux.Handle("/logs/", http.StripPrefix("/logs/", AllowCors(http.HandlerFunc(c.Logs))))
	mux.Handle("/channels", AllowCors(http.HandlerFunc(c.GetChannels)))

	// listen on all localhost
	ip := "127.0.0.1"
	server := &http.Server{Addr: ip + ":" + rPort, Handler: mux}

	// Server listens on its own goroutine
	go func() {
		log.Printf("Listening on %s:%s...", ip, rPort)
		if err := server.ListenAndServe(); err != nil {
			log.Panicf("Listen and serve returned error: %v", err)
		}
	}()

	// Graceful shutdown!

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown returned error %v", err)
	}
}

// AllowCors -
// CORS middleware
func AllowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, req)
	})
}
