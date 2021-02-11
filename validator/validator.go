package validator

import (
	"errors"
	"github.com/asaskevich/govalidator"
	"strings"
)

var (
	// ErrInvalidUsernameFormat gives error message when user attempts to join a lobby when is already in lobby
	ErrInvalidUsernameFormat = errors.New("username is not valid (example valid: 2 < length < 20 and must be A-z 0-9 - _)")
)

// Username must contain alphanumerics, dashes, or unserscores and is from 2 to 20 characters long
func Username(input string) error {
	// if 2 < input < 20, then pass
	if len(input) < 2 || len(input) > 20 {
		return ErrInvalidUsernameFormat
	}

	if !govalidator.IsAlphanumeric(input) {
		if !strings.Contains(input, "-") {
			if !strings.Contains(input, "_") {
				return ErrInvalidUsernameFormat
			}
		}
	}
	return nil
}
