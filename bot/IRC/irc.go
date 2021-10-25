package IRC

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"sync"
	"unicode/utf8"
)

const minpasswordlength = 6

type service struct {
	connection net.Conn
	reader     *textproto.Reader
	writer     *textproto.Writer
	Channels   map[string]struct{}
	m          sync.RWMutex
}

// NewService -
// ignore returns unexported type linter warning (revive)
// nolint:revive
func NewService() (*service, error) {
	return &service{
		Channels: map[string]struct{}{},
	}, nil
}

// expose these as package globals to enable themt o be faked for testing
var tlsLoadX509KeyPair = tls.LoadX509KeyPair
var tlsDial = tls.Dial
var netDial = net.Dial

// Connect to the supplied server, in the correct mode
func (s *service) Connect(server string, useTLS bool) error {
	var err error
	if server == "" {
		return fmt.Errorf("no server supplied, cannot connect to nothing")
	}
	if useTLS {
		cert, certErr := tlsLoadX509KeyPair("cert.pem", "key.pem")
		if certErr != nil {
			log.Printf("error during certificate loading: %v", certErr)
			return certErr
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}
		config.Rand = rand.Reader
		s.connection, err = tlsDial("tcp", server, &config)
	} else {
		s.connection, err = netDial("tcp", server)
	}
	if err != nil {
		log.Printf("dial server using address %s produced error %v", server, err)
		return err
	}
	// Create reader and writer so we can communicate with the server
	r := bufio.NewReader(s.connection)
	w := bufio.NewWriter(s.connection)
	s.reader = textproto.NewReader(r)
	s.writer = textproto.NewWriter(w)

	return nil
}

// Disconnect from the server
func (s *service) Disconnect() error {
	if err := s.writer.PrintfLine("QUIT"); err != nil {
		return fmt.Errorf("disconnect quit error %w", err)
	}
	if err := s.connection.Close(); err != nil {
		return fmt.Errorf("disconnect close error %w", err)
	}
	log.Println("connection close")
	return nil
}

// Login to the server with the supplied credentials
func (s *service) Login(username, password string) error {
	if username == "" {
		return fmt.Errorf("no username supplied for Login, cannot continue")
	}
	if utf8.RuneCountInString(password) < minpasswordlength {
		return fmt.Errorf("password supplied not long enough, got %d, require %d", utf8.RuneCountInString(password), minpasswordlength)
	}
	err := s.writer.PrintfLine("USER %s 8 * :%s", username, username)
	if err != nil {
		return fmt.Errorf("login User error %w", err)
	}
	err = s.writer.PrintfLine("NICK %s", username)
	if err != nil {
		return fmt.Errorf("login Nick error %w", err)
	}
	str := fmt.Sprintf("PRIVMSG NickServ :identify %s %s", username, password)
	log.Printf("nick: %q password: '******' identify", username)
	err = s.writer.PrintfLine(str)
	if err != nil {
		return fmt.Errorf("login identify error %w", err)
	}
	return nil
}

// Join the supplied channel - it doesn't matter if we join the same channel a
// trillion times.
func (s *service) Join(channel string) error {
	if channel == "" {
		// Bail if no channel supplied - it's not an error though
		log.Printf("No channel name to join supplied")
		return nil
	}
	log.Printf("Join channel %s", channel)
	if err := s.writer.PrintfLine(fmt.Sprintf("JOIN %s", channel)); err != nil {
		return fmt.Errorf("channel join error %w", err)
	}

	// Add the channel to the map of channels that the bot has a presence in
	s.m.Lock()
	defer s.m.Unlock()
	s.Channels[channel] = struct{}{}
	return nil
}

// Part from the supplied channel
func (s *service) Part(channel string) error {
	if channel == "" {
		// Bail if no channel supplied - it's not an error though
		log.Printf("No channel name to part supplied")
		return nil
	}
	log.Printf("Part channel %s", channel)
	if err := s.writer.PrintfLine(fmt.Sprintf("PART %s", channel)); err != nil {
		return fmt.Errorf("channel part error %w", err)
	}

	// Remove the channel from the map of channels that the bot has a presence in
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.Channels, channel)
	return nil
}

// Say the supplied text to the supplied channel
func (s *service) Say(text, channel string) error {
	return fmt.Errorf("not implemented")
}
