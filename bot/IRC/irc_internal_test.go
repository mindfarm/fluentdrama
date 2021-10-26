package IRC

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	testcases := map[string]struct {
		server  string
		useTLS  bool
		outErr  error
		cert    tls.Certificate
		certErr error
		dialErr error
	}{
		"No servername supplied": {
			outErr: fmt.Errorf("no server supplied, cannot connect to nothing"),
		},
		"Successful non-tls connection": {
			server: "Fake-server-name",
		},
		"Unsuccessful non-tls connection": {
			server:  "Fake-server-name",
			dialErr: fmt.Errorf("fake dial error"),
			outErr:  fmt.Errorf("fake dial error"),
		},
		"certificate error when creating tls connection": {
			server:  "Fake-server-name",
			useTLS:  true,
			certErr: fmt.Errorf("fake cert error"),
			outErr:  fmt.Errorf("fake cert error"),
		},
		"Successful tls connection": {
			server: "Fake-server-name",
			useTLS: true,
			cert:   tls.Certificate{},
		},
		"Unsuccessful tls connection": {
			server:  "Fake-server-name",
			useTLS:  true,
			dialErr: fmt.Errorf("fake dial error"),
			outErr:  fmt.Errorf("fake dial error"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// Fake functions for testing
			netDial = func(network, address string) (net.Conn, error) {
				// the net.Conn will always be nil because we don't need to
				// check if one actually formed
				return nil, tc.dialErr
			}
			defer func() { netDial = net.Dial }()
			tlsDial = func(network, addr string, config *tls.Config) (*tls.Conn, error) {
				// the tls.Conn will always be nil because we don't need to
				// check if one actually formed
				return nil, tc.dialErr
			}
			defer func() { tlsDial = tls.Dial }()
			tlsLoadX509KeyPair = func(certFile, keyFile string) (tls.Certificate, error) {
				return tc.cert, tc.certErr
			}
			defer func() { tlsLoadX509KeyPair = tls.LoadX509KeyPair }()

			s, _ := NewService()
			err := s.Connect(tc.server, tc.useTLS)
			if tc.outErr == nil {
				assert.Nil(t, err, "got unexpected err %v", err)
				assert.NotNil(t, s.reader, "no reader created")
				assert.NotNil(t, s.writer, "no writer created")
			} else {
				assert.NotNil(t, err, "got nil err, but was expecting %v", tc.outErr)
				assert.EqualError(t, tc.outErr, err.Error())
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	testcases := map[string]struct {
		closeErr error
		writeErr error
		outErr   error
	}{
		"quit error": {
			writeErr: fmt.Errorf("fake-write-error"),
			outErr:   fmt.Errorf("disconnect quit error fake-write-error"),
		},
		"close error": {
			closeErr: fmt.Errorf("fake-close-error"),
			outErr:   fmt.Errorf("disconnect close error fake-close-error"),
		},
		"successful closure": {},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			s.connection = &fakeConn{}
			s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
			closeErr = tc.closeErr
			writeErr = tc.writeErr
			err := s.Disconnect()
			if tc.outErr == nil {
				assert.Nil(t, err, "got unexpected err %v", err)
			} else {
				assert.NotNil(t, err, "got nil err, but was expecting %v", tc.outErr)
				assert.EqualError(t, tc.outErr, err.Error())
			}
		})
	}
}

func TestLogin(t *testing.T) {
	testcases := map[string]struct {
		username string
		password string
		written  []string
		writeErr error
		outErr   error
	}{
		"No username": {
			outErr: fmt.Errorf("no username supplied for Login, cannot continue"),
		},
		"No password": {
			username: "fake-user",
			outErr:   fmt.Errorf("password supplied not long enough, got %d, require %d", 0, minpasswordlength),
		},
		"password too short": {
			username: "fake-user",
			password: "small",
			outErr:   fmt.Errorf("password supplied not long enough, got %d, require %d", utf8.RuneCountInString("small"), minpasswordlength),
		},
		"writer error": {
			username: "fake-user",
			password: "fake-pass",
			writeErr: fmt.Errorf("fake-error"),
			outErr:   fmt.Errorf("login USER error fake-error"),
		},
		"successful login": {
			username: "fake-user",
			password: "fake-pass",
			written:  []string{"USER fake-user 8 * :fake-user\r\n", "NICK fake-user\r\n", "PRIVMSG NickServ : identify fake-user fake-pass\r\n"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			s.reader = textproto.NewReader(bufio.NewReader(&fakeConn{}))
			s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
			writeHold = []string{}
			writeErr = tc.writeErr
			err := s.Login(tc.username, tc.password)
			if tc.outErr == nil {
				assert.Nil(t, err, "got unexpected err %v", err)
				assert.Len(t, writeHold, len(tc.written), "got different length array")
				for i := range tc.written {
					// order of values must be the same, and must be equal
					assert.Equal(t, tc.written[i], writeHold[i])
				}
			} else {
				assert.NotNil(t, err, "got nil err, but was expecting %v", tc.outErr)
				assert.EqualError(t, err, tc.outErr.Error())
			}
		})
	}
}

func TestJoin(t *testing.T) {
	testcases := map[string]struct {
		channels         []string
		expectedChannels []string
		writeErr         error
		outErr           error
	}{
		"No channel name supplied": {
			channels: []string{""},
		},
		"Join error": {
			channels: []string{"fake-channel"},
			writeErr: fmt.Errorf("fake-write-error"),
			outErr:   fmt.Errorf("channel join error fake-write-error"),
		},
		"Single channel": {
			channels:         []string{"fake-channel"},
			expectedChannels: []string{"fake-channel"},
		},
		"Multiple channels with duplicate": {
			channels:         []string{"fake-channel", "second-fake-channel", "fake-channel"},
			expectedChannels: []string{"fake-channel", "second-fake-channel"},
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			s.reader = textproto.NewReader(bufio.NewReader(&fakeConn{}))
			s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
			writeErr = tc.writeErr
			var err error
			for _, c := range tc.channels {
				err = s.Join(c)
			}
			if tc.outErr == nil {
				assert.Nil(t, err, "got unexpected err %v", err)
				assert.Len(t, s.Channels, len(tc.expectedChannels), "got different length array")
				for _, c := range tc.expectedChannels {
					_, ok := s.Channels[c]
					assert.True(t, ok, "missing %s", c)
				}
			} else {
				assert.NotNil(t, err, "got nil err, but was expecting %v", tc.outErr)
				assert.EqualError(t, err, tc.outErr.Error())
			}
		})
	}
}

func TestPart(t *testing.T) {
	testcases := map[string]struct {
		channels         []string
		toDelete         []string
		expectedChannels []string
		writeErr         error
		outErr           error
	}{
		"No channel name supplied": {
			channels:         []string{"fake-channel"},
			expectedChannels: []string{"fake-channel"},
			toDelete:         []string{""},
		},
		"Part error": {
			channels: []string{"fake-channel"},
			toDelete: []string{"fake-channel"},
			writeErr: fmt.Errorf("fake-write-error"),
			outErr:   fmt.Errorf("channel part error fake-write-error"),
		},
		"Single channel": {
			channels: []string{"fake-channel"},
			toDelete: []string{"fake-channel"},
		},
		"Multiple channels with duplicate": {
			toDelete: []string{"fake-channel", "second-fake-channel", "fake-channel"},
			channels: []string{"fake-channel", "second-fake-channel"},
		},
		"Multiple channels with remainder": {
			toDelete:         []string{"fake-channel", "second-fake-channel"},
			channels:         []string{"fake-channel", "second-fake-channel", "third-fake-channel"},
			expectedChannels: []string{"third-fake-channel"},
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			s.reader = textproto.NewReader(bufio.NewReader(&fakeConn{}))
			s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
			// Set up
			writeErr = nil
			for _, c := range tc.channels {
				_ = s.Join(c)
			}

			// Test
			writeErr = tc.writeErr
			var err error
			for _, c := range tc.toDelete {
				err = s.Part(c)
			}

			if tc.outErr == nil {
				assert.Nil(t, err, "got unexpected err %v", err)
				assert.Len(t, s.Channels, len(tc.expectedChannels), "got different length array")
				for _, c := range tc.expectedChannels {
					_, ok := s.Channels[c]
					assert.True(t, ok, "missing %s", c)
				}
			} else {
				assert.NotNil(t, err, "got nil err, but was expecting %v", tc.outErr)
				assert.EqualError(t, err, tc.outErr.Error())
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	testcases := map[string]struct {
		input     string
		prefix    string
		command   string
		trailing  string
		cmdParams string
	}{
		"no colon prefix": {
			input:     "PING :zirconium.libera.chat",
			prefix:    "",
			command:   "PING",
			trailing:  "zirconium.libera.chat",
			cmdParams: "",
		},
		"colon prefix": {
			input:     ":zirconium.libera.chat 376 loggingbot :End of /MOTD command.",
			prefix:    "zirconium.libera.chat",
			command:   "376",
			trailing:  "End of /MOTD command.",
			cmdParams: "loggingbot",
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			output := s.parseline(tc.input)
			assert.Equal(t, tc.prefix, output[0])
			assert.Equal(t, tc.command, output[1])
			assert.Equal(t, tc.trailing, output[2])
			assert.Equal(t, tc.cmdParams, output[3])
		})
	}
}

func TestProcessLine(t *testing.T) {
	testcases := map[string]struct {
		input    string
		writeErr error
	}{
		"ping": {
			input: "PING :zirconium.libera.chat",
		},
		"376": {
			input: ":zirconium.libera.chat 376 loggingbot :End of /MOTD command.",
		},
		"privmsg": {
			input: ":b0nn!~shane@user/b0nn PRIVMSG loggingbot :8",
		},
		"channel message": {},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewService()
			s.reader = textproto.NewReader(bufio.NewReader(&fakeConn{}))
			s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
			writeErr = nil

			// Test
			writeErr = tc.writeErr
			s.processLine(tc.input)
		})
	}
}
