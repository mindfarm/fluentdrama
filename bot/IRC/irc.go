package IRC

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"unicode/utf8"
)

const minpasswordlength = 6

type service struct {
	connection net.Conn
	reader     *textproto.Reader
	writer     *textproto.Writer
}

// NewService -
// ignore returns unexported type linter warning (revive)
// nolint:revive
func NewService() (*service, error) {
	return &service{}, nil
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
func (s *service) Disconnect(server, mode string) error {
	return fmt.Errorf("not implemented")
}

// Login to the server with the supplied credentials
func (s *service) Login(username, password string) error {
	if username == "" {
		return fmt.Errorf("no username supplised for Login, cannot continue")
	}
	if utf8.RuneCountInString(password) < minpasswordlength {
		return fmt.Errorf("password supplied not long enough, got %d, require %d", utf8.RuneCountInString(password), minpasswordlength)
	}
	s.writer.PrintfLine("USER %s 8 * :%s", username, username)
	s.writer.PrintfLine("NICK %s", username)
	str := fmt.Sprintf("PRIVMSG NickServ :identify %s %s", username, password)
	log.Printf("nick: %q password: '******' identify", username)
	s.writer.PrintfLine(str)
	return nil
}

// Join the supplied channel
func (s *service) Join(channel string) error {
	return fmt.Errorf("not implemented")
}

// Part from the supplied channel
func (s *service) Part(channel string) error {
	return fmt.Errorf("not implemented")
}

// Say the supplied text to the supplied channel
func (s *service) Say(output, channel string) error {
	return fmt.Errorf("not implemented")
}
