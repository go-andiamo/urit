package urit

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
)

func TestPathPart_Match_SingleFixed(t *testing.T) {
	vars := &pathVars{}
	pt := &pathPart{
		fixed:      true,
		fixedValue: `foo`,
	}
	require.True(t, pt.match(`foo`, 0, vars, nil, nil))
	require.False(t, pt.match(`bar`, 0, vars, nil, nil))
}

func TestPathPart_Match_SingleVar(t *testing.T) {
	vars := newPathVars(Names)
	pt := &pathPart{
		fixed: false,
		name:  "foo",
	}
	require.True(t, pt.match(`bar`, 0, vars, nil, nil))
	require.Equal(t, 1, vars.Len())
	v, ok := vars.Get(0)
	require.True(t, ok)
	require.Equal(t, `bar`, v)
	v, ok = vars.Get("foo")
	require.True(t, ok)
	require.Equal(t, "bar", v)
}

func TestPathPart_Match_SingleVarRegex(t *testing.T) {
	vars := newPathVars(Names)
	pt := &pathPart{
		fixed:  false,
		name:   "foo",
		regexp: regexp.MustCompile(`^[a-z]{3}$`),
	}
	require.True(t, pt.match(`bar`, 0, vars, nil, nil))
	require.Equal(t, 1, vars.Len())
	v, ok := vars.Get(0)
	require.True(t, ok)
	require.Equal(t, `bar`, v)
	v, ok = vars.Get("foo")
	require.True(t, ok)
	require.Equal(t, "bar", v)

	require.False(t, pt.match(`123`, 0, vars, nil, nil))
}

func TestPathPart_OverallRegexp(t *testing.T) {
	pt := &pathPart{
		subParts: []pathPart{
			{
				fixed:      true,
				fixedValue: `--`,
			},
			{
				fixed:     false,
				orgRegexp: `^(?:[a-z\+\-]{3})$`,
				name:      `foo`,
			},
			{
				fixed:      true,
				fixedValue: `++`,
			},
			{
				fixed:     false,
				orgRegexp: `([0-9]*)?`,
				name:      `bar`,
			},
		},
	}
	rx := pt.overallRegexp()
	require.NotNil(t, rx)

	ss := rx.FindStringSubmatch(`--a+z++12345`)
	require.NotEmpty(t, ss)
	ss = rx.FindStringSubmatch(`--a+z++`)
	require.NotEmpty(t, ss)

	vars := newPathVars(Positions)
	ok := pt.multiMatch(`--a+z++12345`, 0, vars, nil)
	require.True(t, ok)
	require.Equal(t, 2, vars.Len())
	v, ok := vars.Get(0)
	require.True(t, ok)
	require.Equal(t, `a+z`, v)
	v, ok = vars.Get("foo")
	require.True(t, ok)
	require.Equal(t, `a+z`, v)
	v, ok = vars.Get(1)
	require.True(t, ok)
	require.Equal(t, `12345`, v)
	v, ok = vars.Get("bar")
	require.True(t, ok)
	require.Equal(t, `12345`, v)

	ok = pt.multiMatch(`--a+z++`, 0, vars, nil)
	require.True(t, ok)
}

func TestPathPart_OverallRegexpMatch(t *testing.T) {
	vars := newPathVars(Positions)
	pt := pathPart{
		subParts: []pathPart{
			{
				fixed:      true,
				fixedValue: `--`,
			},
			{
				fixed:     false,
				orgRegexp: ``,
				name:      `foo`,
			},
			{
				fixed:      true,
				fixedValue: `++`,
			},
			{
				fixed:     false,
				orgRegexp: ``,
				name:      `bar`,
			},
		},
	}

	ok := pt.multiMatch(`--abc++123`, 0, vars, nil)
	require.True(t, ok)
	require.Equal(t, 2, vars.Len())
	v, ok := vars.Get(0)
	require.True(t, ok)
	require.Equal(t, "abc", v)
	v, ok = vars.Get(1)
	require.True(t, ok)
	require.Equal(t, "123", v)

	vars.Clear()
	ok = pt.multiMatch(`--abc++123`, 0, vars, []VarMatchOption{&fooUpperChecker{}})
	require.True(t, ok)
	require.Equal(t, 2, vars.Len())
	v, ok = vars.Get(0)
	require.True(t, ok)
	require.Equal(t, "ABC", v)
	v, ok = vars.Get(1)
	require.True(t, ok)
	require.Equal(t, "123", v)
}

type fooUpperChecker struct {
}

func (o *fooUpperChecker) Match(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool) {
	return strings.ToUpper(value), true
}
func (o *fooUpperChecker) Applicable(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) bool {
	return name == "foo"
}
