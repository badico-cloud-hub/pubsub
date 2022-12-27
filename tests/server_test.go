package tests

import (
	"testing"

	"github.com/badico-cloud-hub/pubsub/api"
)

func TestNewServer(t *testing.T) {
	port := "3000"
	server := api.NewServer(port)
	if server == nil {
		t.Errorf("TestNewServer: expect(!=nil) - got(nil)\n")
	}
}
