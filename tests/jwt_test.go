package tests

import (
	"testing"

	"github.com/badico-cloud-hub/pubsub/adapters"
	"github.com/joho/godotenv"
)

//go test --run TestNewJWT
func TestNewJWT(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestNewJWT: expect(nil) - got(%s)\n", err.Error())
	}
	jwtAdapter := adapters.NewJWTAdapter()
	if jwtAdapter == nil {
		t.Errorf("TestNewJWT: expect(!nil) - got(%v)\n", jwtAdapter)
	}
}

//go test --run TestVerifyToken
func TestVerifyToken(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestVerifyToken: expect(nil) - got(%s)\n", err.Error())
	}
	jwtAdapter := adapters.NewJWTAdapter()
	if jwtAdapter == nil {
		t.Errorf("TestVerifyToken: expect(!nil) - got(%v)\n", jwtAdapter)
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhc3NvY2lhdGlvbl9pZCI6InBpeEBvcmlvbiIsImNsaWVudF9pZCI6IjAwMDAwMUBvcmlvbiIsImV4cCI6MTY4NzU2NzEzNCwiaWF0IjoxNjg3NTY3MTA0LCJuYmYiOjE2ODc1NjcxMDQsInNjb3BlcyI6bnVsbH0.GeB88kOOVXcFSYXBdoD_4XhvXiKFSOm3kK9EG3ZeS4U"
	_, err := jwtAdapter.VerifyToken(token)
	if err != nil {
		t.Errorf("TestVerifyToken: expect(nil) - got(%s)\n", err.Error())
	}
}
