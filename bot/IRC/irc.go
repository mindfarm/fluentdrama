package IRC

import "fmt"

type service struct {
}

// NewService -
// ignore returns unexported type linter warning (revive)
// nolint:revive
func NewService() (*service, error) {
	return &service{}, nil
}

// Connect to the supplied server, in the correct mode
func (s *service) Connect(server, mode string) error {
	return fmt.Errorf("not implemented")
}

// Disconnect from the server
func (s *service) Disconnect(server, mode string) error {
	return fmt.Errorf("not implemented")
}

// Login to the server with the supplied credentials
func (s *service) Login(username, password string) error {
	return fmt.Errorf("not implemented")
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
