package IRC_test

import (
	"fmt"
	"testing"

	"github.com/mindfarm/fluentdrama/bot/IRC"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	testcases := map[string]struct {
		outChan  chan []byte
		owner    string
		outError error
	}{
		"Happy path": {
			owner:   "fake-owner",
			outChan: make(chan []byte),
		},
		"No owner": {
			outChan:  make(chan []byte),
			outError: fmt.Errorf("no owner supplied"),
		},
		"No out channel": {
			owner:    "fake-owner",
			outError: fmt.Errorf("no out channel supplied"),
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			service, err := IRC.NewService(tc.owner, tc.outChan)
			if tc.outError == nil {
				// no error expected, but a service is
				assert.Nil(t, err, "No error expected, but got %v", err)
				assert.NotNil(t, service, "Expected an instance of service, but got nil instead")
			} else {
				// no service expected, but an error is
				assert.NotNil(t, err, "Expected error %v, but got nil instead", tc.outError)
				assert.EqualError(t, err, tc.outError.Error(), "expected %v, got %v", tc.outError, err)
				assert.Nil(t, service, "No service expected, but got %v", err)
			}
		})
	}
}
