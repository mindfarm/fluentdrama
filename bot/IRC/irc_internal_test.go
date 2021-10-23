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

var readHold []byte

type fakeConn struct{}

func (f *fakeConn) Read(p []byte) (int, error) {
	readHold = p
	return 0, nil
}

var writeHold []string

func (f *fakeConn) Write(p []byte) (int, error) {
	// convert p to a string so that it will persist
	// if it's left as a []byte it will be overwritten because of the semantics
	// of the bufio writer using a pointer :'(
	writeHold = append(writeHold, string(p))
	//
	return len(p), nil
}

func TestLogin(t *testing.T) {
	s, _ := NewService()
	s.reader = textproto.NewReader(bufio.NewReader(&fakeConn{}))
	s.writer = textproto.NewWriter(bufio.NewWriter(&fakeConn{}))
	testcases := map[string]struct {
		username string
		password string
		written  []string
		outErr   error
	}{
		"No username": {
			outErr: fmt.Errorf("no username supplised for Login, cannot continue"),
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
		"successful login": {
			username: "fake-user",
			password: "fake-pass",
			written:  []string{"USER fake-user 8 * :fake-user\r\n", "NICK fake-user\r\n", "PRIVMSG NickServ :identify fake-user fake-pass\r\n"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			defer func() { fmt.Println("EMPTYING"); writeHold = []string{} }()
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
				assert.EqualError(t, tc.outErr, err.Error())
			}
		})
	}
}
