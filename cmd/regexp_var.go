package cmd

import (
	"fmt"
	"regexp"
)

type MatchAny struct{}

func (MatchAny) Match(string) bool { return true }

type Regexp struct{ *regexp.Regexp }

func (r Regexp) String() string {
	if r.Regexp == nil {
		return fmt.Sprintf("%s", (*regexp.Regexp)(nil))
	}
	return r.Regexp.String()
}
func (r *Regexp) Set(from string) (err error) {
	if len(from) == 0 {
		return nil
	}
	r.Regexp, err = regexp.Compile(from)
	return err
}
func (r Regexp) Type() string         { return "Regexp" }
func (r Regexp) Match(it string) bool { return r.MatchString(it) }

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type Value interface {
	String() string
	Set(string) error
	Type() string
}
