package domain

import (
	"net/http"
)

type StatusCode uint8

const (
	StatusOK      StatusCode = 0
	ErrNotFound   StatusCode = 1
	ErrBadRequest StatusCode = 2
	ErrInternal   StatusCode = 3
)

type StatusMessage struct {
	Code    int
	Name    string
	Message string
}

var ErrorToCodes = []StatusMessage{
	{Code: http.StatusOK, Name: "StatusOK", Message: "Success!"},
	{Code: http.StatusNotFound, Name: "ErrNotFound", Message: "No receipt found for that ID."},
	{Code: http.StatusBadRequest, Name: "ErrBadRequest", Message: "The receipt is invalid."},
	{Code: http.StatusInternalServerError, Name: "ErrInternalServer", Message: "Internal services have failed"},
}
