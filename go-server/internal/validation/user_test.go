package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestUserName(t *testing.T) {
	maxName := strings.Repeat("я", MaxNameLength)
	cases := []struct {
		in      string
		want    string
		wantErr error
	}{
		{"Alice", "Alice", nil},
		{"  Алиса \t", "Алиса", nil},
		{maxName, maxName, nil},
		{"", "", ErrNameEmpty},
		{"   ", "", ErrNameEmpty},
		{maxName + "я", "", ErrNameTooLong},
	}
	for _, c := range cases {
		got, err := UserName(c.in)
		if got != c.want || !errors.Is(err, c.wantErr) {
			t.Errorf("UserName(%q) = %q, %v; want %q, %v", c.in, got, err, c.want, c.wantErr)
		}
	}
}
