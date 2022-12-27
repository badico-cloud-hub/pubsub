package utils

import (
	"crypto"
	"encoding/hex"
)

func GenerateApiKey(text string) (string, error) {
	hash := crypto.SHA256.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
