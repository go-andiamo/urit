package urit

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

type fx struct{}

func (o *fx) Match(value string, expected string, pathPos int, vars PathVars) bool {
	return true
}

type vr struct{}

func (o *vr) Match(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool) {
	return value, true
}

func (o *vr) Applicable(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) bool {
	return true
}

func TestOptionType(t *testing.T) {
	var f interface{} = &fx{}
	var v interface{} = &vr{}

	_, isF := f.(FixedMatchOption)
	require.True(t, isF)
	_, isV := f.(VarMatchOption)
	require.False(t, isV)

	_, isF = v.(FixedMatchOption)
	require.False(t, isF)
	_, isV = v.(VarMatchOption)
	require.True(t, isV)
}
