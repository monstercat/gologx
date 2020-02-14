package server

import (
	"errors"

	"github.com/Masterminds/squirrel"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	TableService = "service"
)


var (
	ErrInvalidId = errors.New("invalid id")
)