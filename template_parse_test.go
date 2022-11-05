package urit

import (
	"fmt"
	"github.com/go-andiamo/splitter"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	tmp, err := NewTemplate("foo/?bar/:bar/{baz}/{qux:[a-z]*}/")
	require.NoError(t, err)
	require.NotNil(t, tmp)
	require.Equal(t, Names, tmp.VarsType())
	require.Equal(t, `/foo/?bar/:bar/{baz}/{qux:[a-z]*}/`, tmp.OriginalTemplate())
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 5, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.Equal(t, `foo`, rt.pathParts[0].fixedValue)
	require.False(t, rt.pathParts[1].fixed)
	require.Nil(t, rt.pathParts[1].regexp)
	require.Equal(t, `bar`, rt.pathParts[1].name)
	require.False(t, rt.pathParts[2].fixed)
	require.Nil(t, rt.pathParts[2].regexp)
	require.Equal(t, `bar`, rt.pathParts[2].name)
	require.False(t, rt.pathParts[3].fixed)
	require.Nil(t, rt.pathParts[3].regexp)
	require.Equal(t, `baz`, rt.pathParts[3].name)
	require.False(t, rt.pathParts[4].fixed)
	require.NotNil(t, rt.pathParts[4].regexp)
	require.Equal(t, `^[a-z]*$`, rt.pathParts[4].regexp.String())
	require.Equal(t, `qux`, rt.pathParts[4].name)

	require.Equal(t, 3, len(rt.namedVars))
	require.Equal(t, 2, len(rt.namedVars["bar"]))
	require.Equal(t, 1, len(rt.namedVars["baz"]))
	require.Equal(t, 1, len(rt.namedVars["qux"]))
}

func TestNewTemplatePositional(t *testing.T) {
	tmp, err := NewTemplate("foo/?/?")
	require.NoError(t, err)
	require.NotNil(t, tmp)
	require.Equal(t, Positions, tmp.VarsType())
	require.Equal(t, `/foo/?/?`, tmp.OriginalTemplate())
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 3, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.Equal(t, `foo`, rt.pathParts[0].fixedValue)
	require.False(t, rt.pathParts[1].fixed)
	require.Nil(t, rt.pathParts[1].regexp)
	require.Equal(t, ``, rt.pathParts[1].name)
	require.False(t, rt.pathParts[2].fixed)
	require.Nil(t, rt.pathParts[2].regexp)
	require.Equal(t, ``, rt.pathParts[2].name)
}

func TestMustCreateTemplatePanics(t *testing.T) {
	const bad = `/foo{`
	_, err := NewTemplate(bad)
	require.Error(t, err)
	require.Panics(t, func() {
		MustCreateTemplate(bad)
	})
	tmp := MustCreateTemplate(`/foo/{bar}`)
	require.NotNil(t, tmp)
	require.Equal(t, `/foo/{bar}`, tmp.OriginalTemplate())
}

func TestNewTemplateWithMultiVarParts(t *testing.T) {
	tmp, err := NewTemplate("{bar}-{baz}")
	require.NoError(t, err)
	require.NotNil(t, tmp)
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 1, len(rt.pathParts))
	require.False(t, rt.pathParts[0].fixed)
	require.Equal(t, 3, len(rt.pathParts[0].subParts))
	require.False(t, rt.pathParts[0].subParts[0].fixed)
	require.Equal(t, "bar", rt.pathParts[0].subParts[0].name)
	require.True(t, rt.pathParts[0].subParts[1].fixed)
	require.Equal(t, "-", rt.pathParts[0].subParts[1].fixedValue)
	require.False(t, rt.pathParts[0].subParts[2].fixed)
	require.Equal(t, "baz", rt.pathParts[0].subParts[2].name)
}

func TestNewTemplateWithQuotes(t *testing.T) {
	tmp, err := NewTemplate(`/foo/"aaa"`)
	require.NoError(t, err)
	require.NotNil(t, tmp)
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 2, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.Equal(t, `foo`, rt.pathParts[0].fixedValue)
	require.True(t, rt.pathParts[1].fixed)
	require.Equal(t, `aaa`, rt.pathParts[1].fixedValue)

	tmp, err = NewTemplate(`/foo/()-"aaa"`)
	require.NoError(t, err)
	require.NotNil(t, tmp)
	rt, ok = tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 2, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.Equal(t, `foo`, rt.pathParts[0].fixedValue)
	require.True(t, rt.pathParts[1].fixed)
	require.Equal(t, `()-aaa`, rt.pathParts[1].fixedValue)
}

func TestNewTemplate_ParseErrors(t *testing.T) {
	testCases := []struct {
		template    string
		expectErr   string
		expectedPos int
	}{
		{
			``,
			`template empty`,
			0,
		},
		{
			`/foo//bar`,
			`path parts cannot be empty`,
			5,
		},
		{
			`/foo/{}`,
			`path var name cannot be empty`,
			5,
		},
		{
			`/foo/{bar:\{[]()}`,
			`path var regexp problem`,
			8,
		},
		{
			`/foo/?/{bar}`,
			`template cannot contain both positional and named path variables`,
			0,
		},
		{
			`/foo/{:[a-z]*}`,
			`path var name cannot be empty`,
			5,
		},
		{
			`/foo/{  :[a-z]*}`,
			`path var name cannot be empty`,
			5,
		},
		{
			`/foo/{foo}/--{}--`,
			`path var name cannot be empty`,
			13,
		},
		{
			`/foo/{foo}/--{bar}--{:[a-z]*}`,
			`path var name cannot be empty`,
			20,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]%s", i+1, tc.template), func(t *testing.T) {
			_, err := NewTemplate(tc.template)
			require.Error(t, err)
			require.Equal(t, tc.expectErr, err.Error())
			sErr, ok := err.(TemplateParseError)
			require.True(t, ok)
			require.Error(t, sErr)
			require.Equal(t, tc.expectedPos, sErr.Position())
		})
	}
}

func TestNewTemplate_MultiArg(t *testing.T) {
	tmp, err := NewTemplate("/foo/--{bar}-{baz}--")
	require.NoError(t, err)
	require.NotNil(t, tmp)
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 2, len(rt.pathParts))
	require.True(t, rt.pathParts[0].fixed)
	require.False(t, rt.pathParts[1].fixed)
	require.Equal(t, 5, len(rt.pathParts[1].subParts))
}

func TestNewTemplate_UnbalancedCurlyErrors(t *testing.T) {
	testCases := []struct {
		str       string
		expectErr string
	}{
		{
			"/foo/{bar",
			"unclosed '{' at position 5",
		},
		{
			"/foo/{bar{",
			"unclosed '{' at position 9",
		},
		{
			"/foo/}bar",
			"unopened '}' at position 5",
		},
		{
			"/foo/{{bar",
			"unclosed '{' at position 6",
		},
		{
			"/foo/{bar}}",
			"unopened '}' at position 10",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]%s", i+1, tc.str), func(t *testing.T) {
			_, err := NewTemplate(tc.str)
			require.Error(t, err)
			require.Equal(t, tc.expectErr, err.Error())
		})
	}
}

func TestTemplate_Sub(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{foo}/`)
	require.NoError(t, err)
	require.NotNil(t, tmp)
	ort, _ := tmp.(*template)

	tmp2, err := tmp.Sub(`{bar}-{baz}/qux`)
	require.NoError(t, err)
	require.Equal(t, `/foo/{foo}/{bar}-{baz}/qux`, tmp2.OriginalTemplate())
	rt, _ := tmp2.(*template)
	require.Equal(t, 4, len(rt.pathParts))
	require.Equal(t, ort.pathParts[0].fixed, rt.pathParts[0].fixed)
	require.Equal(t, ort.pathParts[1].fixed, rt.pathParts[1].fixed)
	require.False(t, rt.pathParts[2].fixed)
	require.True(t, rt.pathParts[3].fixed)

	tmp, err = NewTemplate(`/foo/{foo}`)
	require.NoError(t, err)
	tmp2, err = tmp.Sub(`{bar}`)
	require.NoError(t, err)
	require.Equal(t, `/foo/{foo}/{bar}`, tmp2.OriginalTemplate())
}

func TestTemplate_SubErrors(t *testing.T) {
	tmp, err := NewTemplate(`/foo/{foo}/`)
	require.NoError(t, err)

	_, err = tmp.Sub(`///`)
	require.Error(t, err)
	require.Equal(t, `path parts cannot be empty`, err.Error())

	_, err = tmp.Sub(`?`)
	require.Error(t, err)
	require.Equal(t, `template cannot contain both positional and named path variables`, err.Error())

	tmp, _ = NewTemplate(`/foo/?/`)
	_, err = tmp.Sub(`{foo}`)
	require.Error(t, err)
	require.Equal(t, `template cannot contain both positional and named path variables`, err.Error())
}

func TestTemplateParseWithSplitterOptions(t *testing.T) {
	xo := &extraSplitterOption{}
	tmp, err := NewTemplate(`/foo/{foo}/`, xo)
	require.NoError(t, err)
	require.Equal(t, 2, xo.called)
	rt, ok := tmp.(*template)
	require.True(t, ok)
	require.Equal(t, 2, len(rt.pathParts))
	require.Equal(t, "FOO", rt.pathParts[0].fixedValue)
}

type extraSplitterOption struct {
	called int
}

func (o *extraSplitterOption) Apply(s string, pos int, totalLen int, captured int, skipped int, isLast bool, subParts ...splitter.SubPart) (string, bool, error) {
	o.called++
	return strings.ToUpper(s), true, nil
}
