package cgbdb

import (
	"errors"

	"github.com/jackc/pgerrcode"
)

const (
	DbErrorUniqueViolation = pgerrcode.UniqueViolation
)

var (
	ErrAccountNotFound = errors.New("account not found")
)
