package urit

import (
	"regexp"
	"strings"
)

type fixedMatchOptions []FixedMatchOption
type varMatchOptions []VarMatchOption

func (opts fixedMatchOptions) check(value string, expected string, pathPos int, vars PathVars) bool {
	ok := false
	for _, o := range opts {
		ok = o.Match(value, expected, pathPos, vars)
		if ok {
			break
		}
	}
	return ok
}

func (opts varMatchOptions) check(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool, bool) {
	ok := false
	result := value
	checked := 0
	for _, o := range opts {
		if o.Applicable(value, position, name, rx, rxs, pathPos, vars) {
			checked++
			if s, oko := o.Match(value, position, name, rx, rxs, pathPos, vars); oko {
				result = s
				ok = oko
				break
			}
		}
	}
	return result, ok, checked > 0
}

type FixedMatchOption interface {
	Match(value string, expected string, pathPos int, vars PathVars) bool
}

type VarMatchOption interface {
	Applicable(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) bool
	Match(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool)
}

type caseInsensitiveFixed struct{}

func (o *caseInsensitiveFixed) Match(value string, expected string, pathPos int, vars PathVars) bool {
	return value == expected || strings.EqualFold(value, expected)
}

var (
	_CaseInsensitiveFixed = &caseInsensitiveFixed{}
)
var (
	CaseInsensitiveFixed = _CaseInsensitiveFixed
)
