package errors

import "errors"

var (
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrNoServersProvided = errors.New("no servers provided")
	ErrTimeout           = errors.New("timeout waiting for servers")
	ErrQuorumNotReached  = errors.New("quorum not reached")
	ErrConnectionFailed  = errors.New("connection failed")
	ErrInvalidPattern    = errors.New("invalid pattern")
)
