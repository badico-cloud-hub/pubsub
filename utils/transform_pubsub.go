package utils

import (
	"encoding/base64"
)

//DecodeBase64PEM return string parsed from base64 string
func DecodeBase64PEM(pemString string) (string, error) {
	bytesPem, err := base64.StdEncoding.DecodeString(pemString)
	if err != nil {
		return "", err
	}
	return string(bytesPem), nil
}
