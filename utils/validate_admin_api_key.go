package utils

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/badico-cloud-hub/pubsub/dto"
)

func ValidateAdminApiKey(apiKey, adminBase64 string) (bool, dto.AdminObject) {
	logger := NewLogger(os.Stdout)
	adminApisKeysJson, err := base64.StdEncoding.DecodeString(adminBase64)

	if err != nil {
		logger.Error(err.Error())
		return false, dto.AdminObject{}
	}

	adminObjects := []dto.AdminObject{}

	if err := json.Unmarshal(adminApisKeysJson, &adminObjects); err != nil {
		logger.Error(err.Error())
		return false, dto.AdminObject{}
	}

	for _, admin := range adminObjects {
		if admin.ApiKey == apiKey {
			return true, admin
		}
	}

	return false, dto.AdminObject{}

}
