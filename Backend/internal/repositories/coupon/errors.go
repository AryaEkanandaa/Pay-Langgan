package coupon

import (
	"errors"
	"strings"
)

var ErrDuplicate = errors.New("duplicate entry")
var ErrUsageLimit = errors.New("coupon usage limit reached")

func isPGUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "23505")
}
