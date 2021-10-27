package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mindfarm/fluentdrama/bot/IRC"
)

func main() {
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

	// Create an instance of the server
	out := make(chan []byte)
	s, err := IRC.NewService(owner, out)
	if err != nil {
		panic(err)
	}

	if err = s.Connect(server, secureBool); err != nil {
		log.Fatalf("Could not connect to server with error %v", err)
	}
	go s.Listen()
	go func() {
		// Just dump the output for now
		for {
			log.Println(string(<-out))
		}
	}()

	// TODO - drop this sleep and call the Login when a 376 is detected (end of
	// motd)
	time.Sleep(5 * time.Second)
	err = s.Login(username, password)
	if err != nil {
		log.Fatalf("Unable to login with the following issue: %v", err)
	}
	time.Sleep(5 * time.Second)

	// Example of telling the bot to join a channel
	channel := "#examplechannel"
	if err := s.Join(channel); err != nil {
		log.Printf("Unable to join channel %s", channel)
	} else {
		if err := s.Say(channel, "This channel is now logged"); err != nil {
			log.Println("error sending join message", err)
		}
	}

	// hold the main thread open forever
	select {}
}
