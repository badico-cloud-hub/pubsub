package adapters

import (
	"os"

	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/golang-jwt/jwe"
)

//JWEAdapter is struct for adapter
type JWEAdapter struct {
	publicKey  string
	privateKey string
}

//NewJWEAdapter return new jwt adapter
func NewJWEAdapter() *JWEAdapter {
	return &JWEAdapter{}
}

func (j *JWEAdapter) Init() error {
	public, err := utils.DecodeBase64PEM(os.Getenv("PUBLIC_PEM"))
	if err != nil {
		return err
	}
	j.publicKey = public
	private, err := utils.DecodeBase64PEM(os.Getenv("PRIVATE_PEM"))
	if err != nil {
		return err
	}
	j.privateKey = private
	return nil
}

func (j *JWEAdapter) Decrypt(token string) (string, error) {
	k, err := jwe.ParseRSAPrivateKeyFromPEM([]byte(j.privateKey))
	if err != nil {
		return "", err
	}

	jtk, err := jwe.ParseEncrypted(token)
	if err != nil {
		return "", err
	}

	decripted, err := jtk.Decrypt(k)
	if err != nil {
		return "", err
	}

	return string(decripted), nil
}
