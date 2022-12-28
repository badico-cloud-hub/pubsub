package tests

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/joho/godotenv"
)

func TestGenerateApiKey(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGenerateApiKey: expect(nil) - got(%s)\n", err.Error())
	}
	secret := os.Getenv("SECRET")
	plainKey := fmt.Sprintf("%s:%s:%s", "id@lucas", "admin", secret)
	apiKey, err := utils.GenerateApiKey(plainKey)
	if err != nil {
		t.Errorf("TestGenerateApiKey: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("api-key: %s\n", apiKey)
}
