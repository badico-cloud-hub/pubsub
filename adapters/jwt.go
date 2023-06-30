package adapters

import (
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnespectedSigninMethod    = errors.New("unespected signin method")
	ErrInvalidKey                = errors.New("key is invalid")
	ErrInvalidKeyType            = errors.New("key is of invalid type")
	ErrHashUnavailable           = errors.New("the requested hash function is unavailable")
	ErrTokenMalformed            = errors.New("token is malformed")
	ErrTokenUnverifiable         = errors.New("token is unverifiable")
	ErrTokenSignatureInvalid     = errors.New("token signature is invalid")
	ErrTokenRequiredClaimMissing = errors.New("token is missing required claim")
	ErrTokenInvalidAudience      = errors.New("token has invalid audience")
	ErrTokenExpired              = errors.New("token is expired")
	ErrTokenUsedBeforeIssued     = errors.New("token used before issued")
	ErrTokenInvalidIssuer        = errors.New("token has invalid issuer")
	ErrTokenInvalidSubject       = errors.New("token has invalid subject")
	ErrTokenNotValidYet          = errors.New("token is not valid yet")
	ErrTokenInvalidId            = errors.New("token has invalid id")
	ErrTokenInvalidClaims        = errors.New("token has invalid claims")
	ErrInvalidType               = errors.New("invalid type for claim")
	ErrTokenInvalid              = errors.New("token invalid")
)

var allErrors = []error{
	ErrUnespectedSigninMethod,
	ErrInvalidKey,
	ErrInvalidKeyType,
	ErrHashUnavailable,
	ErrTokenMalformed,
	ErrTokenUnverifiable,
	ErrTokenSignatureInvalid,
	ErrTokenRequiredClaimMissing,
	ErrTokenInvalidAudience,
	ErrTokenExpired,
	ErrTokenUsedBeforeIssued,
	ErrTokenInvalidIssuer,
	ErrTokenInvalidSubject,
	ErrTokenNotValidYet,
	ErrTokenInvalidId,
	ErrTokenInvalidClaims,
	ErrInvalidType,
	ErrTokenInvalid,
}

//JWTAdapter is struct for adapter
type JWTAdapter struct {
	secret string
}

//NewJWTAdapter return new jwt adapter
func NewJWTAdapter() *JWTAdapter {
	return &JWTAdapter{
		secret: os.Getenv("SECRET_AUTH"),
	}
}

func (j *JWTAdapter) verifyError(err error) error {
	for _, e := range allErrors {
		if strings.Contains(err.Error(), e.Error()) {
			return e
		}
	}
	return ErrTokenInvalid
}

//VerifyToken verify token signature
func (j *JWTAdapter) VerifyToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnespectedSigninMethod
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return jwt.MapClaims{}, j.verifyError(err)
	}

	return token.Claims, nil
}
