package config

import "regexp"

// Regexp is a type that supports JSON Unmarshalling
type Regexp struct {
	regexp.Regexp
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
// also used by json.Unmarshal.
func (r *Regexp) UnmarshalJSON(text []byte) error {
	rr, err := regexp.Compile(string(text))
	if err != nil {
		return err
	}
	*r = Regexp{*rr}
	return nil
}
