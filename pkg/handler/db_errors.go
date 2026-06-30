package handler

import (
	"strings"
)

// flagKeyUniqueViolation reports whether err is a DB unique violation on flags.key (idx_flag_key).
func flagKeyUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") &&
		(strings.Contains(msg, "idx_flag_key") || strings.Contains(msg, "flags.key") || strings.Contains(msg, "key"))
}