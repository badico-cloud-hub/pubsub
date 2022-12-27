package infra

import "errors"

var (
	ErrorServiceNotFound          = errors.New("service not found")
	ErrorServiceAlreadyExist      = errors.New("service already exist")
	ErrorServiceEventNotFound     = errors.New("service event not found")
	ErrorServiceEventAlreadyExist = errors.New("service event already exist")
	ErrorClientNotFound           = errors.New("client not found")
	ErrorClientAlreadyExist       = errors.New("client already exist")
	ErrorSubscriptinEventNotFound = errors.New("not subscription event in table")
)
