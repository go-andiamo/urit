package urit

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathVars_Get(t *testing.T) {
	args := Positional("a", "b")

	v, ok := args.Get(0)
	require.True(t, ok)
	require.Equal(t, "a", v)
	v, ok = args.Get(1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.Get(-1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.Get(-2)
	require.True(t, ok)
	require.Equal(t, "a", v)

	all := args.GetAll()
	require.Equal(t, 2, len(all))
	require.Equal(t, "a", all[0].Value)
	require.Equal(t, "b", all[1].Value)

	_, ok = args.Get(2)
	require.False(t, ok)
	_, ok = args.Get(-3)
	require.False(t, ok)
	args.Clear()
	_, ok = args.Get(0)
	require.False(t, ok)

	_, ok = args.Get(2, true)
	require.False(t, ok)
	_, ok = args.Get(false, true)
	require.False(t, ok)

}

func TestPositional(t *testing.T) {
	args := Positional("a", "b")
	require.Equal(t, Positions, args.VarsType())
	require.Equal(t, 2, args.Len())
	all := args.GetAll()
	require.Equal(t, 2, len(all))
	require.Equal(t, "a", all[0].Value)
	require.Equal(t, 0, all[0].Position)
	require.Equal(t, "", all[0].Name)
	require.Equal(t, "b", all[1].Value)
	require.Equal(t, 1, all[1].Position)
	require.Equal(t, "", all[1].Name)

	err := args.AddPositionalValue("c")
	require.NoError(t, err)
	require.Equal(t, 3, args.Len())
	err = args.AddNamedValue("foo", "bar")
	require.Error(t, err)
}

func TestNamed(t *testing.T) {
	args := Named("foo", "a", "bar", "b", "bar", "c")
	require.Equal(t, Names, args.VarsType())
	require.Equal(t, 3, args.Len())
	all := args.GetAll()
	require.Equal(t, 3, len(all))
	require.Equal(t, "foo", all[0].Name)
	require.Equal(t, 0, all[0].Position)
	require.Equal(t, "a", all[0].Value)
	require.Equal(t, 0, all[0].NamedPosition)

	require.Equal(t, "bar", all[1].Name)
	require.Equal(t, 1, all[1].Position)
	require.Equal(t, "b", all[1].Value)
	require.Equal(t, 0, all[1].NamedPosition)

	require.Equal(t, "bar", all[2].Name)
	require.Equal(t, 2, all[2].Position)
	require.Equal(t, "c", all[2].Value)
	require.Equal(t, 1, all[2].NamedPosition)

	err := args.AddNamedValue("qux", "d")
	require.NoError(t, err)
	require.Equal(t, 4, args.Len())
	err = args.AddPositionalValue("whoops")
	require.Error(t, err)
}

func TestNamedPanics(t *testing.T) {
	require.Panics(t, func() {
		Named("a", "b", "c") // not an even number!
	})
	require.Panics(t, func() {
		Named(true, false) // first should be a string!
	})
}

func TestPathVars_Get_Positional(t *testing.T) {
	args := Positional("a", "b")
	require.Equal(t, 2, args.Len())
	v, ok := args.GetPositional(0)
	require.True(t, ok)
	require.Equal(t, "a", v)
	v, ok = args.GetPositional(1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.GetPositional(-1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	_, ok = args.GetPositional(2)
	require.False(t, ok)

	v, ok = args.Get(0)
	require.True(t, ok)
	require.Equal(t, "a", v)
	v, ok = args.Get(1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.Get(-1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	_, ok = args.Get(2)
	require.False(t, ok)
	_, ok = args.Get("foo")
	require.False(t, ok)
	_, ok = args.Get("foo", 0)
	require.False(t, ok)

	_, ok = args.GetNamed("foo", 0)
	require.False(t, ok)
}

func TestPathVars_Get_Named(t *testing.T) {
	args := Named("foo", "a", "bar", "b", "bar", "c", "bar", "d")
	require.Equal(t, 4, args.Len())

	v, ok := args.GetNamed("foo", 0)
	require.True(t, ok)
	require.Equal(t, "a", v)
	_, ok = args.GetNamed("foo", 1)
	require.False(t, ok)
	v, ok = args.GetNamed("bar", 0)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.GetNamedFirst("bar")
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.GetNamed("bar", 1)
	require.True(t, ok)
	require.Equal(t, "c", v)
	v, ok = args.GetNamedLast("bar")
	require.True(t, ok)
	require.Equal(t, "d", v)
	v, ok = args.GetNamed("bar", -1)
	require.True(t, ok)
	require.Equal(t, "d", v)
	v, ok = args.GetNamed("bar", 2)
	require.True(t, ok)
	require.Equal(t, "d", v)
	_, ok = args.GetNamed("bar", 3)
	require.False(t, ok)
	_, ok = args.GetNamed("xxx", 0)
	require.False(t, ok)
	_, ok = args.GetNamedFirst("xxx")
	require.False(t, ok)
	_, ok = args.GetNamedLast("xxx")
	require.False(t, ok)

	v, ok = args.GetPositional(0)
	require.True(t, ok)
	require.Equal(t, "a", v)
	v, ok = args.GetPositional(1)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.GetPositional(2)
	require.True(t, ok)
	require.Equal(t, "c", v)
	v, ok = args.GetPositional(-1)
	require.True(t, ok)
	require.Equal(t, "d", v)
	v, ok = args.GetPositional(-2)
	require.True(t, ok)
	require.Equal(t, "c", v)
	v, ok = args.GetPositional(-3)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.GetPositional(3)
	require.True(t, ok)
	require.Equal(t, "d", v)
	_, ok = args.GetPositional(4)
	require.False(t, ok)

	v, ok = args.Get("foo")
	require.True(t, ok)
	require.Equal(t, "a", v)
	_, ok = args.Get("foo", 1)
	require.False(t, ok)
	v, ok = args.Get("foo", -1)
	require.True(t, ok)
	require.Equal(t, "a", v)

	v, ok = args.Get("bar")
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.Get("bar", 0)
	require.True(t, ok)
	require.Equal(t, "b", v)
	v, ok = args.Get("bar", 1)
	require.True(t, ok)
	require.Equal(t, "c", v)
}
