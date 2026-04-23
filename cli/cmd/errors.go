package cmd

import "strings"

func isRBACDeniedError(err error) bool {
	message := strings.TrimSpace(strings.ToLower(err.Error()))
	return strings.Contains(message, "access to ") && strings.HasSuffix(message, " is denied")
}
