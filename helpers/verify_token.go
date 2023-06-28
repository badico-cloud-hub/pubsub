package helpers

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/badico-cloud-hub/pubsub/adapters"
	"github.com/badico-cloud-hub/pubsub/dto"
)

var (
	ErrTokenBadFormated        = errors.New("token bad formated")
	ErrTokenPayloadBadFormated = errors.New("token payload bad formated")
)

//ValidatePayloadJwtDto execute validation for payload the token
func ValidatePayloadJwtDto(jwtDto dto.JWTDTO) error {
	if jwtDto.Payload.ClientId == "" || jwtDto.Payload.AssociationId == "" || jwtDto.Payload.ApiKeyType == "" || jwtDto.Payload.Provider == "" {
		return ErrTokenPayloadBadFormated
	}

	return nil
}

//VerifyToken execute validation for bearer token
func VerifyToken(bearerToken string) (dto.JWTDTO, error) {
	jweAdapter := adapters.NewJWEAdapter()
	if err := jweAdapter.Init(); err != nil {
		return dto.JWTDTO{}, err
	}
	jwtAdapter := adapters.NewJWTAdapter()

	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ToUpper(parts[0]) != "BEARER" {
		return dto.JWTDTO{}, ErrTokenBadFormated
	}
	decripted, err := jweAdapter.Decrypt(parts[1])
	if err != nil {
		return dto.JWTDTO{}, err
	}
	claims, err := jwtAdapter.VerifyToken(decripted)
	if err != nil {
		return dto.JWTDTO{}, err
	}

	bCl, err := json.Marshal(claims)
	if err != nil {
		return dto.JWTDTO{}, err
	}
	jwtDto := dto.JWTDTO{}
	if err := json.Unmarshal(bCl, &jwtDto); err != nil {
		return dto.JWTDTO{}, err
	}
	if err := ValidatePayloadJwtDto(jwtDto); err != nil {
		return dto.JWTDTO{}, err
	}
	return jwtDto, nil
}
