package query

import "errors"

var (
	ErrTimeout = errors.New("query step timeout")

	ErrPlanNotFound = errors.New("query plan not found")

	ErrInvalidItem = errors.New("invalid item")

	ErrHashNotFound = errors.New("hash not found")
)
