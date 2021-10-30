package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mindfarm/fluentdrama/bot/IRC"
	data "github.com/mindfarm/fluentdrama/bot/repository/postgres"
)

func main() {
	dbURI, ok := os.LookupEnv("DBURI")
	if !ok {
		log.Fatalf("DBURI is not set")
	}

	owner, ok := os.LookupEnv("BOT_OWNER")
	if !ok {
		log.Fatal("env var BOT_OWNER not set, cannot continue")
	}

	server, ok := os.LookupEnv("IRC_SERVER")
	if !ok {
		log.Fatal("env var IRC_SERVER not set, cannot continue")
	}

	secureStr, ok := os.LookupEnv("SECURE")
	if !ok {
		log.Print("env var SECURE not set, defaulting to `false`")
		secureStr = "false"
	}
	secureBool, err := strconv.ParseBool(secureStr)
	if err != nil {
		log.Fatalf("env var SECURE was not a valid boolean, please use `true` or `false`, got %q", secureStr)
	}

	username, ok := os.LookupEnv("IRC_USERNAME")
	if !ok {
		log.Fatal("env var IRC_USERNAME not set, cannot continue")
	}

	password, ok := os.LookupEnv("IRC_PASSWORD")
	if !ok {
		log.Fatal("env var IRC_PASSWORD not set, cannot continue")
	}

	// Datastore
	ds, err := data.NewPgCustomerRepo(dbURI)
	if err != nil {
		log.Fatalf("Unable to connect to datastore with error %v", err)
	}
	fmt.Println(ds)
	channels, err := ds.GetChannels(context.Background())
	if err != nil {
		log.Printf("error fetching channels %v", err)
	}

	// Create an instance of the server
	out := make(chan map[string]string)
	s, err := IRC.NewService(owner, channels, out)
	if err != nil {
		panic(err)
	}

	// Connect to the server
	if err = s.Connect(server, secureBool); err != nil {
		log.Fatalf("Could not connect to server with error %v", err)
	}

	go s.Listen()
	// m := map[string]string{}
	go func() {
		// Just dump the output for now
		for {
			m := <-out
			log.Printf("%#v", m)
			c, ok := m["Command"]
			if !ok {
				log.Printf("No command detected in %#v", m)
				continue
			}
			switch strings.ToUpper(c) {
			case "JOIN":
				if p, ok := m["Prefix"]; ok && strings.Split(p, "!")[0] == username {
					if err = ds.AddChannel(context.Background(), m["CmdParams"]); err != nil {
						log.Printf("Error adding channel %s %v", c, err)
					} else {
						log.Println("Successfully added channel ", c)
					}
				}
			case "PRIVMSG":
				// log channel messagesAddLog(ctx context.Context, channel, username, said string) error
				// ignore private messages sent to the bot
				if cp, ok := m["CmdParams"]; ok && strings.Split(cp, "!")[0] != username {
					if err = ds.AddLog(context.Background(), m["CmdParams"], strings.Split(m["Prefix"], "!")[0], m["Trailing"]); err != nil {
						log.Printf("Error adding log %#v %v", m, err)
					} else {
						log.Println("Successfully added log ", m)
					}
				}
			}
		}
	}()

	// TODO - drop this sleep and call the Login when a 376 is detected (end of
	// motd)
	time.Sleep(5 * time.Second)
	err = s.Login(username, password)
	if err != nil {
		log.Fatalf("Unable to login with the following issue: %v", err)
	}
	time.Sleep(15 * time.Second)
	// hold the main thread open forever
	select {}
}
