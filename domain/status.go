package domain

import (
	"net/http"
)

type StatusCode uint8

const (
	StatusOK      StatusCode = 0
	ErrNotFound   StatusCode = 1
	ErrBadRequest StatusCode = 2
	ErrConflict   StatusCode = 3
	ErrInternal   StatusCode = 4
)

type StatusMessage struct {
	Code    int
	Message string
}

var ErrorToCodes = []StatusMessage{
	{Code: http.StatusOK, Message: "Success!"},
	{Code: http.StatusNotFound, Message: "No receipt found for that ID."},
	{Code: http.StatusBadRequest, Message: "The receipt is invalid."},
}
