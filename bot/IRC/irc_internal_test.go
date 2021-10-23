package IRC

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"testing"

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
