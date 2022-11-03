package urit

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"regexp"
	"testing"
)

func TestMatchString(t *testing.T) {
	tmp, err := NewTemplate("/foo/{foo}/bar/{bar}/{bar}")
	require.NoError(t, err)
	require.NotNil(t, tmp)

	_, ok := tmp.Matches("/foo/fooey/bar/aaa/")
	require.False(t, ok)

	args, ok := tmp.Matches("/foo/fooey/bar/aaa/bbb")
	require.True(t, ok)
	v, ok := args.GetPositional(0)
	require.True(t, ok)
	require.Equal(t, "fooey", v)
	v, ok = args.GetPositional(1)
	require.True(t, ok)
	require.Equal(t, "aaa", v)
	v, ok = args.GetPositional(2)
	require.True(t, ok)
	require.Equal(t, "bbb", v)

	v, ok = args.GetNamedFirst("foo")
	require.True(t, ok)
	require.Equal(t, "fooey", v)
	v, ok = args.GetNamedFirst("bar")
	require.True(t, ok)
	require.Equal(t, "aaa", v)
	v, ok = args.GetNamedLast("bar")
	require.True(t, ok)
	require.Equal(t, "bbb", v)
}

func TestMatchPositional(t *testing.T) {
	tmp, err := NewTemplate("/foo/?/bar/?/?")
	require.NoError(t, err)
	require.NotNil(t, tmp)

	_, ok := tmp.Matches("/foo/fooey/bar/aaa/")
	require.False(t, ok)

	args, ok := tmp.Matches("/foo/fooey/bar/aaa/bbb")
	require.True(t, ok)
	v, ok := args.GetPositional(0)
	require.True(t, ok)
	require.Equal(t, "fooey", v)
	v, ok = args.GetPositional(1)
	require.True(t, ok)
	require.Equal(t, "aaa", v)
	v, ok = args.GetPositional(2)
	require.True(t, ok)
	require.Equal(t, "bbb", v)
	_, ok = args.GetPositional(3)
	require.False(t, ok)
}

func TestMatchStringMulti(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{foo}/bar/--{bar: [a-zA-Z]*}-{bar: [0-9]*}--`)
	require.NoError(t, err)
	require.NotNil(t, tmp)

	args, ok := tmp.Matches(`/foo/fooey/bar/--abc-123--`)
	require.True(t, ok)
	v, ok := args.GetPositional(0)
	require.True(t, ok)
	require.Equal(t, "fooey", v)
	v, ok = args.GetPositional(1)
	require.True(t, ok)
	require.Equal(t, "abc", v)
	v, ok = args.GetPositional(2)
	require.True(t, ok)
	require.Equal(t, "123", v)

	v, ok = args.GetNamedFirst("foo")
	require.True(t, ok)
	require.Equal(t, "fooey", v)
	v, ok = args.GetNamedFirst("bar")
	require.True(t, ok)
	require.Equal(t, "abc", v)
	v, ok = args.GetNamedLast("bar")
	require.True(t, ok)
	require.Equal(t, "123", v)

	req, err := http.NewRequest("GET", `https://www.example.com/foo/fooey/bar/--abc-123--`, nil)
	require.NoError(t, err)
	_, ok = tmp.MatchesRequest(req)
	require.True(t, ok)

	_, ok = tmp.Matches(`/foo/fooey/bar/--abc+123--`)
	require.False(t, ok)
}

func TestTemplate_PathFrom(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{foo}/bar/--{bar: [a-zA-Z]*}-{bar: [0-9]*}--`)
	require.NoError(t, err)

	pth, err := tmp.PathFrom(Named(
		"foo", "fooey",
		"bar", "abc",
		"bar", "123"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/--abc-123--`, pth)

	_, err = tmp.PathFrom(Named(
		"foo", "fooey"))
	require.Error(t, err)
	require.Equal(t, `no var for 'bar'`, err.Error())

	_, err = tmp.PathFrom(Named(
		"foo", "fooey",
		"bar", "abc"))
	require.Error(t, err)
	require.Equal(t, `no var for 'bar' (position 2)`, err.Error())
}

func TestTemplate_PathFrom_Positional(t *testing.T) {
	tmp, err := NewTemplate(`/foo/?/bar/?/baz/?`)
	require.NoError(t, err)

	pth, err := tmp.PathFrom(Positional("fooey", "barey", "bazey"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/barey/baz/bazey`, pth)

	_, err = tmp.PathFrom(Positional("fooey", "barey"))
	require.Error(t, err)
	require.Equal(t, `no var for position 3`, err.Error())

	_, err = tmp.PathFrom(Positional("fooey"))
	require.Error(t, err)
	require.Equal(t, `no var for position 2`, err.Error())

	_, err = tmp.PathFrom(Positional())
	require.Error(t, err)
	require.Equal(t, `no var for position 1`, err.Error())
}

func TestTemplate_ResolveTo(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{foo:[a-z]*}/bar/--{bar: [a-zA-Z]*}-{bar: [0-9]*}--`)
	require.NoError(t, err)

	tmp2, err := tmp.ResolveTo(Named(
		"foo", "fooey",
		"bar", "abc"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/--abc-{bar:[0-9]*}--`, tmp2.OriginalTemplate())
	rt, ok := tmp2.(*template)
	require.True(t, ok)
	require.NotNil(t, rt)
	require.Equal(t, 4, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.True(t, rt.pathParts[1].fixed)
	require.True(t, rt.pathParts[2].fixed)
	require.False(t, rt.pathParts[3].fixed)
	require.Equal(t, 5, len(rt.pathParts[3].subParts))
	require.True(t, rt.pathParts[3].subParts[0].fixed)
	require.True(t, rt.pathParts[3].subParts[1].fixed)
	require.True(t, rt.pathParts[3].subParts[2].fixed)
	require.False(t, rt.pathParts[3].subParts[3].fixed)
	require.True(t, rt.pathParts[3].subParts[4].fixed)

	tmp2, err = tmp.ResolveTo(Named(
		"foo", "fooey",
		"bar", "abc",
		"bar", "345"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/--abc-345--`, tmp2.OriginalTemplate())

	tmp2, err = tmp.ResolveTo(Named("bar", "abc"))
	require.NoError(t, err)
	require.Equal(t, `/foo/{foo:[a-z]*}/bar/--abc-{bar:[0-9]*}--`, tmp2.OriginalTemplate())
}

func TestTemplate_ResolveTo_Positional(t *testing.T) {
	tmp, err := NewTemplate(`/foo/?/bar/?/baz/?`)
	require.NoError(t, err)

	tmp2, err := tmp.ResolveTo(Positional())
	require.NoError(t, err)
	require.Equal(t, `/foo/?/bar/?/baz/?`, tmp2.OriginalTemplate())

	tmp2, err = tmp.ResolveTo(Positional("fooey"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/?/baz/?`, tmp2.OriginalTemplate())

	tmp2, err = tmp.ResolveTo(Positional("fooey", "barey"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/barey/baz/?`, tmp2.OriginalTemplate())
	_, err = tmp2.PathFrom(Positional())
	require.Error(t, err)

	tmp2, err = tmp.ResolveTo(Positional("fooey", "barey", "bazey"))
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/barey/baz/bazey`, tmp2.OriginalTemplate())
	str, err := tmp2.PathFrom(Positional())
	require.NoError(t, err)
	require.Equal(t, `/foo/fooey/bar/barey/baz/bazey`, str)
}

func TestCaseInsensitiveFixed_Match(t *testing.T) {
	tmp, err := NewTemplate(`/foo/?/bar`)
	require.NoError(t, err)

	_, ok := tmp.Matches(`Foo/123/Bar`)
	require.False(t, ok)
	_, ok = tmp.Matches(`Foo/123/Bar`, CaseInsensitiveFixed)
	require.True(t, ok)
}

func TestTemplate_Matches_WithVarOption(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{id1:uuid4}/bar/{id2:[0-9]*}`)
	require.NoError(t, err)

	_, ok := tmp.Matches(`foo/febb16e3-0827-46bd-abed-f88ed9cc35ff/bar/123`)
	require.False(t, ok)

	_, ok = tmp.Matches(`foo/febb16e3-0827-46bd-abed-f88ed9cc35ff/bar/123`, &uuidChecker{})
	require.True(t, ok)

	_, ok = tmp.Matches(`foo/febb16e3-0827-46bd-abed-f88ed9cc35ff/bar/abc`, &uuidChecker{})
	require.False(t, ok)

	tmp, err = NewTemplate(`/foo/{id1:uuid4}/bar/{id2:[0-9]*}`, CaseInsensitiveFixed)
	require.NoError(t, err)
	_, ok = tmp.Matches(`FOO/febb16e3-0827-46bd-abed-f88ed9cc35ff/BAR/123`)
	require.False(t, ok)
	_, ok = tmp.Matches(`FOO/febb16e3-0827-46bd-abed-f88ed9cc35ff/BAR/abc`, &uuidChecker{})
	require.False(t, ok)
	_, ok = tmp.Matches(`FOO/febb16e3-0827-46bd-abed-f88ed9cc35ff/BAR/123`, &uuidChecker{})
	require.True(t, ok)
}

func TestTemplate_MergeOptions(t *testing.T) {
	testCases := []struct {
		initOptions  []interface{}
		addOptions   []interface{}
		expectFixeds int
		expectVars   int
	}{
		{
			[]interface{}{},
			[]interface{}{},
			0,
			0,
		},
		{
			[]interface{}{CaseInsensitiveFixed},
			[]interface{}{},
			1,
			0,
		},
		{
			[]interface{}{CaseInsensitiveFixed},
			[]interface{}{CaseInsensitiveFixed},
			1,
			0,
		},
		{
			[]interface{}{},
			[]interface{}{CaseInsensitiveFixed},
			1,
			0,
		},
		{
			[]interface{}{&uuidChecker{}},
			[]interface{}{},
			0,
			1,
		},
		{
			[]interface{}{&uuidChecker{}},
			[]interface{}{&uuidChecker{}},
			0,
			1,
		},
		{
			[]interface{}{},
			[]interface{}{&uuidChecker{}},
			0,
			1,
		},
		{
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			[]interface{}{},
			1,
			1,
		},
		{
			[]interface{}{},
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			1,
			1,
		},
		{
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			[]interface{}{&uuidChecker{}},
			1,
			2,
		},
		{
			[]interface{}{&uuidChecker{}},
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			1,
			2,
		},
		{
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			[]interface{}{&uuidChecker{}, dummyVar{}, dummyFixed{}},
			1,
			2,
		},
		{
			[]interface{}{&uuidChecker{}},
			[]interface{}{&dummyVar{}, &dummyFixed{}},
			1,
			2,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			tmp, err := NewTemplate(`/foo/{id1:uuid4}`, tc.initOptions...)
			require.NoError(t, err)
			rt, ok := tmp.(*template)
			require.True(t, ok)
			fs, vs := rt.mergeOptions(tc.addOptions)
			require.Equal(t, tc.expectFixeds, len(fs))
			require.Equal(t, tc.expectVars, len(vs))
		})
	}
}

type uuidChecker struct {
}

var uuid4Rx = regexp.MustCompile("^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-4[0-9A-Fa-f]{3}-[89abAB][0-9A-Fa-f]{3}-[0-9A-Fa-f]{12}$")

func (o *uuidChecker) Match(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool) {
	return value, uuid4Rx.MatchString(value)
}
func (o *uuidChecker) Applicable(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) bool {
	return rxs == "uuid4"
}

type dummyFixed struct {
}

func (o *dummyFixed) Match(value string, expected string, pathPos int, vars PathVars) bool {
	return true
}

type dummyVar struct {
}

func (o *dummyVar) Match(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) (string, bool) {
	return value, true
}
func (o *dummyVar) Applicable(value string, position int, name string, rx *regexp.Regexp, rxs string, pathPos int, vars PathVars) bool {
	return true
}
