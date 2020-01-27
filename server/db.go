package server

import (
	"errors"

	"github.com/Masterminds/squirrel"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	TableOrigin = "origin"

	ViewService = "service_view"
)


var (
	ErrInvalidId = errors.New("invalid id")
)