// Package validation checks and normalizes user input.
package validation

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxNameLength = 100

var (
	ErrNameEmpty   = errors.New("'name' must be a non-empty string")
	ErrNameTooLong = errors.New("'name' must be at most " + strconv.Itoa(MaxNameLength) + " characters")
)

// UserName trims surrounding whitespace and checks the length in characters.
func UserName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrNameEmpty
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return "", ErrNameTooLong
	}
	return name, nil
}
