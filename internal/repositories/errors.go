package repositories

import (
	"errors"
	"strings"
)

var ErrDuplicate = errors.New("duplicate entry")

func isPGUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "23505")
}
