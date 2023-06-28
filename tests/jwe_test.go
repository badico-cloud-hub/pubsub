package tests

import (
	"fmt"
	"testing"

	"github.com/badico-cloud-hub/pubsub/adapters"
	"github.com/joho/godotenv"
)

//go test --run TestNewJWE
func TestNewJWE(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestNewJWE: expect(nil) - got(%s)\n", err.Error())
	}
	jwtAdapter := adapters.NewJWEAdapter()
	if jwtAdapter == nil {
		t.Errorf("TestNewJWE: expect(!nil) - got(%v)\n", jwtAdapter)
	}
}

//go test --run TestDecryptToken
func TestDecryptToken(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestDecryptToken: expect(nil) - got(%s)\n", err.Error())
	}
	jweAdapter := adapters.NewJWEAdapter()
	if jweAdapter == nil {
		t.Errorf("TestDecryptToken: expect(!nil) - got(%v)\n", jweAdapter)
	}
	if err := jweAdapter.Init(); err != nil {
		t.Errorf("TestDecryptToken: expect(nil) - got(%s)\n", err.Error())
	}
	token := "eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZHQ00ifQ.wgLfq8uGhbGJliYp6u2L5CBvoBEvso7rsM8aK5Pd_QJsTzaJZRteqKf__tXKAAaklB3aS5w6urHRPANcAmA6nBt5OEj6Uumq-p46BSRzIBiJVXAqdLI85S6QyqGXqnLo7fsPYNyom-RRKm0VI6F3OpDwwuOv_ebAAmfV4ILWknGTV19ylyBifnbO0zogmkACA8rOVy9PucywfSmNhw__dKYXx24DUyLasnBNfvZe16xBmqk3vdzQqo3FBNQnOmNxuc5Xz1IT_AWdMLPVIJsdEVOa66fCk-OM55mO2YK_yXcwXPtlyIKGoPbqcD69_JaOuozz0km2-xRx9Bndt2Ha_vij5HZ9FPVJTN6GbkxLyz28-ov8j2JUNQNo1rawdKFE4bLSm4zmUv9Jr_lnrVPu3fuyOxtV8orQM0HgK1bveyeS_5FndFECLMJzKnO_tbrf0qD9jLmIEac16bS82SMRM2dm7mskpM4Sg680Q5fcmjPAKCtzpdn-x-J0h1xkxY87.8SLpHO6HkMkmuMtb._fq2gc6_Sa9iItdCVz_a2mPLscjzU3ilN8TrmitMPlKmGfIB1xSlK7dPizgZrJnMA0kBRjjj6xEXw5Ixoz7EguxEL0N_3kN8alitc76LtUCg_ANK0cG-Em_JISaLUuN0whQhTZWaqGwCqVXvfzr_-qp8BvpzwzFMkbuMBhv1DCD4KYmeX6JAhKRXRA6x9wHBkpMUWXM73Qkh0p6Uta-uJXhvBQQBz6GzY2hr5__D5xUKoJvOTbTTna8piKeaIMbONE6-Ra628KdAUOEzOhRi1fVPNNg3hh4V-DkxmTMq0LJz244AUi_RlXGNUqMEJPQuzJGZQQ.Dca1spzcdLWaw9X8zunAnA"
	decrypted, err := jweAdapter.Decrypt(token)
	if err != nil {
		t.Errorf("TestDecryptToken: expect(nil) - got(%s)\n", err.Error())
	}
	fmt.Printf("decripted: %s\n", decrypted)
}
