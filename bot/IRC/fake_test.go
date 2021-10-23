package IRC

import (
	"net"
	"time"
)

// This file holds faked functions used in unit tests

var readHold []byte

type fakeAddr struct{}

func (fa *fakeAddr) Network() string {
	return ""
}

func (fa *fakeAddr) String() string {
	return ""
}

type fakeConn struct{}

func (f *fakeConn) Read(p []byte) (int, error) {
	readHold = p
	return 0, nil
}

var writeHold []string
var writeErr error

func (f *fakeConn) Write(p []byte) (int, error) {
	writeHold = append(writeHold, string(p))
	return len(p), writeErr
}

var closeErr error

func (f *fakeConn) Close() error {
	return closeErr
}

func (f *fakeConn) LocalAddr() net.Addr {
	return &fakeAddr{}
}

func (f *fakeConn) RemoteAddr() net.Addr {
	return &fakeAddr{}
}

func (f *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (f *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
