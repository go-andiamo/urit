package urit

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewHeaders(t *testing.T) {
	h, err := NewHeaders()
	require.NoError(t, err)
	require.NotNil(t, h)
	rh, ok := h.(*headers)
	require.True(t, ok)
	require.Equal(t, 0, len(rh.entries))

	h, err = NewHeaders("foo", 1.23)
	require.NoError(t, err)
	require.NotNil(t, h)
	rh, ok = h.(*headers)
	require.True(t, ok)
	require.Equal(t, 1, len(rh.entries))
	require.Equal(t, 1.23, rh.entries["foo"])
}

func TestNewHeadersErrors(t *testing.T) {
	_, err := NewHeaders("foo") // must be an even number!
	require.Error(t, err)
	require.Equal(t, `must be a value for each name`, err.Error())

	_, err = NewHeaders(true, false) // first must be a string!
	require.Error(t, err)
	require.Equal(t, `name must be a string`, err.Error())
}

func TestHeaders_GetHeaders(t *testing.T) {
	h, err := NewHeaders()
	require.NoError(t, err)
	hds, err := h.GetHeaders()
	require.NoError(t, err)
	require.Equal(t, 0, len(hds))

	h, err = NewHeaders("foo", 1.23)
	require.NoError(t, err)
	hds, err = h.GetHeaders()
	require.NoError(t, err)
	require.Equal(t, 1, len(hds))
	require.Equal(t, "1.23", hds["foo"])

	h, err = NewHeaders("foo", nil)
	require.NoError(t, err)
	_, err = h.GetHeaders()
	require.Error(t, err)
	require.Equal(t, `unknown value type`, err.Error())

	h, err = NewHeaders("foo", func() {
		// this does not yield a string
	})
	require.NoError(t, err)
	_, err = h.GetHeaders()
	require.Error(t, err)
	require.Equal(t, `unknown value type`, err.Error())
}

func TestHeaders_GetSet(t *testing.T) {
	h, err := NewHeaders("foo", 1)
	require.NoError(t, err)
	v, ok := h.Get("foo")
	require.True(t, ok)
	require.Equal(t, 1, v)
	h.Set("foo", 2)
	v, ok = h.Get("foo")
	require.True(t, ok)
	require.Equal(t, 2, v)
	_, ok = h.Get("bar")
	require.False(t, ok)
}

func TestHeaders_HasDel(t *testing.T) {
	h, err := NewHeaders("foo", 1)
	require.NoError(t, err)
	require.True(t, h.Has("foo"))
	h.Del("foo")
	require.False(t, h.Has("foo"))
	require.False(t, h.Has("bar"))
}

func TestHeaders_Clone(t *testing.T) {
	h1, err := NewHeaders("foo", 1)
	require.NoError(t, err)
	require.True(t, h1.Has("foo"))
	require.False(t, h1.Has("bar"))
	h2 := h1.Clone()
	require.True(t, h2.Has("foo"))
	require.False(t, h2.Has("bar"))
	h2.Del("foo")
	require.True(t, h1.Has("foo"))
	require.False(t, h1.Has("bar"))
	require.False(t, h2.Has("foo"))
	require.False(t, h2.Has("bar"))
	h2.Set("bar", 2)
	require.True(t, h1.Has("foo"))
	require.False(t, h1.Has("bar"))
	require.False(t, h2.Has("foo"))
	require.True(t, h2.Has("bar"))
}
