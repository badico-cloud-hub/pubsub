package infra

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

//SecretManagerClient is
type SecretManagerClient struct {
	Environments map[string]interface{}
}

var instance *SecretManagerClient

func NewSecretManagerClient() *SecretManagerClient {
	if instance == nil {
		instance = &SecretManagerClient{
			Environments: make(map[string]interface{}),
		}
		instance.load()
	}
	return instance
}

//Get return value of variable
func (s *SecretManagerClient) Get(envVar string) interface{} {
	value := s.Environments[envVar]
	return value
}

//load execute loading the envs
func (s *SecretManagerClient) load() {
	enviroments := os.Environ()
	for _, line := range enviroments {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "{SECRETSMANAGER:") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			secretParts := s.GetValueBetweenString(parts[1], "{SECRETSMANAGER:", "}")
			if secretParts != "" {
				secretPartsArray := strings.Split(secretParts, ":")
				secretType := secretPartsArray[0]
				secretName := secretPartsArray[len(secretPartsArray)-1]
				secretData := s.RetrieveSecret(secretName)

				if secretType == "SIMPLE" {
					var parsedData map[string]interface{}
					err := json.Unmarshal([]byte(secretData), &parsedData)
					if err != nil {
						log.Fatal(err)
					}

					stValue := parsedData["value"].(string)
					s.Environments[key] = stValue
				} else if secretType == "COMPOSED" {
					secretDataParts := secretPartsArray[1 : len(secretPartsArray)-1]
					var parsedData map[string]interface{}
					err := json.Unmarshal([]byte(secretData), &parsedData)
					if err != nil {
						log.Fatal(err)
					}

					stValue := parsedData["value"].(string)
					secretValues := strings.Split(stValue, ":")
					for index, part := range secretDataParts {
						s.Environments[fmt.Sprintf("%s_%s", key, part)] = secretValues[index]
					}
				} else if secretType == "RDSDB" {
					var dbConfig map[string]interface{}
					err := json.Unmarshal([]byte(secretData), &dbConfig)
					if err != nil {
						log.Fatal(err)
					}

					for field, value := range dbConfig {
						upperCaseField := strings.ToUpper(field)
						s.Environments[fmt.Sprintf("%s_%s", key, upperCaseField)] = value
					}
				} else if secretType == "DINAMIC" {
					var parsedData map[string]interface{}
					err := json.Unmarshal([]byte(secretData), &parsedData)
					if err != nil {
						log.Fatal(err)
					}
					s.Environments[key] = parsedData
				}
			}
		} else {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := parts[1]
			s.Environments[key] = value
		}
	}
}

//getValueBetweenString return value into string env
func (s *SecretManagerClient) GetValueBetweenString(haystack, start, end string) string {
	startPos := strings.Index(haystack, start)
	if startPos == -1 {
		return ""
	}

	endPos := strings.Index(haystack[startPos+len(start):], end)
	if endPos == -1 {
		return ""
	}

	return haystack[startPos+len(start) : startPos+len(start)+endPos]
}

//retrieveSecret configure secret manager
func (s *SecretManagerClient) RetrieveSecret(secretName string) string {
	svc := secretsmanager.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
	})))

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Fatal(err)
	}

	return *result.SecretString
}
