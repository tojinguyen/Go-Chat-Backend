package error

import "errors"

var (
	ErrKeyDoesNotExist = errors.New("key does not exist")
	ErrTokenInvalid    = errors.New("token is invalid")
	ErrTokenExpired    = errors.New("token is expired")
)
