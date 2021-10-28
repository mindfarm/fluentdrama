package IRC

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"unicode/utf8"
)

const minpasswordlength = 6

type service struct {
	connection net.Conn
	reader     *textproto.Reader
	writer     *textproto.Writer
	Channels   map[string]struct{}
	out        chan []byte
	m          sync.RWMutex
	Username   string
	Owner      string
}

// NewService -
// ignore returns unexported type linter warning (revive)
// nolint:revive
func NewService(owner string, out chan []byte) (*service, error) {
	if owner == "" {
		return nil, fmt.Errorf("no owner supplied")
	}
	if out == nil {
		return nil, fmt.Errorf("no out channel supplied")
	}
	return &service{
		Channels: map[string]struct{}{},
		Owner:    owner,
		out:      out,
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

	if err := s.writer.PrintfLine("USER %s 8 * :%s", username, username); err != nil {
		return fmt.Errorf("login USER error %w", err)
	}

	if err := s.writer.PrintfLine("NICK %s", username); err != nil {
		return fmt.Errorf("login NICK error %w", err)
	}

	s.Username = username
	authStr := fmt.Sprintf("PRIVMSG NickServ : identify %s %s", username, password)
	if err := s.writer.PrintfLine(authStr); err != nil {
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
func (s *service) Say(target, text string) error {
	if target == "" {
		return fmt.Errorf("say has no target supplied")
	}
	if text == "" {
		return fmt.Errorf("say has no text supplied")
	}
	s2 := fmt.Sprintf("%s %s : %s", "PRIVMSG", target, text)
	if err := s.writer.PrintfLine(s2); err != nil {
		return fmt.Errorf("cannot say %s to %s because error %w", text, target, err)
	}
	log.Printf("Say %s to %s", text, target)
	return nil
}

func (s *service) Listen() {
	for {
		line, err := s.reader.ReadLine()
		if err != nil {
			log.Printf("Error reading socket %v", err)
			log.Fatal()
		}
		s.processLine(line)
	}
}

func (s *service) processLine(line string) {
	parsed := s.parseline(line)
	switch parsed[1] {
	case "376":
		// 376 is the end of the MOTD
		for c := range s.Channels {
			if err := s.Join(c); err != nil {
				log.Printf("Error joining channel %q, %v", c, err)
			}
		}
	case "PING":
		out := strings.Replace(line, "PING", "PONG", -1)
		if err := s.writer.PrintfLine(out); err != nil {
			log.Printf("Error %v when writing %s", err, out)
		}
	case "PRIVMSG", "JOIN":
		// messages directed at the bot
		if parsed[3] == s.Username {
			if parsed[0] == s.Owner {
				// Commands from the owner
				str := strings.Split(parsed[2], " ")
				switch strings.ToLower(str[0]) {
				case "join":
					if err := s.Join(str[1]); err != nil {
						log.Println("join error", err)
					}
				case "part":
					if err := s.Part(str[1]); err != nil {
						log.Println("part error", err)
					}
				case "say":
					if err := s.Say(str[1], strings.Join(str[2:], " ")); err != nil {
						log.Println("say error", err)
					}
				}
			} else {
				// pass the message on to the owner
				owner := strings.Split(parsed[0], "!")
				if err := s.Say(owner[0], fmt.Sprint("Got this message ", parsed[2])); err != nil {
					log.Printf("error %v", err)
				}
			}
		} else {
			log.Println("Putting message onto channel")
			log.Println(line)
			data, err := json.Marshal(
				map[string]string{
					"Prefix":    parsed[0],
					"Command":   parsed[1],
					"Trailing":  parsed[2],
					"CmdParams": parsed[3],
				})
			if err != nil {
				log.Printf("Error marshalling parsed %#v %v", parsed, err)
			}
			s.out <- data
		}
	}
}

func (s *service) parseline(line string) []string {
	prefixEnd := -1
	var Prefix, Command, Trailing, CmdParams string

	if strings.HasPrefix(line, ":") {
		prefixEnd = strings.Index(line, " ")
		Prefix = line[1:prefixEnd]
	}

	trailingStart := strings.Index(line, " :")
	if trailingStart >= 0 {
		Trailing = line[trailingStart+2:]
	} else {
		trailingStart = len(line) - 1
	}

	cmdAndParams := strings.Fields(line[(prefixEnd + 1) : trailingStart+1])
	if len(cmdAndParams) > 0 {
		Command = cmdAndParams[0]
	}

	if len(cmdAndParams) > 1 {
		CmdParams = strings.Join(cmdAndParams[1:], "")
	}

	fmt.Println(Prefix, Command, Trailing, CmdParams)
	return []string{Prefix, Command, Trailing, CmdParams}
}
