package utils

import "errors"

var ErrInvalidToken = errors.New("invalid token")
var ErrExpiredToken = errors.New("token has expired")
var ErrNoRowsInserted = errors.New("no rows were inserted")
var ErrUnauthorized = errors.New("unauthorized")
var ErrValueConversion = errors.New("could not convert value")
