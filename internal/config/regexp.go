package config

import (
	"regexp"
	"strings"
)

// Regexp is a type that supports JSON Unmarshalling
type Regexp struct {
	regexp.Regexp
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (r *Regexp) UnmarshalJSON(text []byte) error {
	rr, err := regexp.Compile(strings.Trim(string(text), `"`))
	if err != nil {
		return err
	}
	*r = Regexp{*rr}
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (r *Regexp) MarshalJSON() ([]byte, error) {
	return []byte(r.Regexp.String()), nil
}
