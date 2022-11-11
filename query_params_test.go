package urit

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewQueryParams(t *testing.T) {
	p, err := NewQueryParams()
	require.NoError(t, err)
	require.NotNil(t, p)
	rp, ok := p.(*queryParams)
	require.True(t, ok)
	require.Equal(t, 0, len(rp.params))
	q, err := p.GetQuery()
	require.NoError(t, err)
	require.Equal(t, ``, q)

	p, err = NewQueryParams("foo", 1.23)
	require.NoError(t, err)
	require.NotNil(t, p)
	rp, ok = p.(*queryParams)
	require.True(t, ok)
	require.Equal(t, 1, len(rp.params))
	require.Equal(t, 1, len(rp.params["foo"]))
	require.Equal(t, 1.23, rp.params["foo"][0])
}

func TestNewQueryParamsErrors(t *testing.T) {
	_, err := NewQueryParams("foo") // must be an even number!
	require.Error(t, err)
	require.Equal(t, `must be a value for each name`, err.Error())

	_, err = NewQueryParams(true, false) // first must be a string!
	require.Error(t, err)
	require.Equal(t, `name must be a string`, err.Error())
}

func TestQueryParams_GetQuery(t *testing.T) {
	p, err := NewQueryParams()
	require.NoError(t, err)
	q, err := p.GetQuery()
	require.NoError(t, err)
	require.Equal(t, ``, q)

	p, err = NewQueryParams("foo", nil)
	require.NoError(t, err)
	q, err = p.GetQuery()
	require.NoError(t, err)
	require.Equal(t, `?foo`, q)

	p, err = NewQueryParams("foo", nil, "foo", true)
	require.NoError(t, err)
	q, err = p.GetQuery()
	require.NoError(t, err)
	require.Contains(t, q, `foo=true`)
	require.True(t, strings.HasPrefix(q, "?"))
	require.Equal(t, strings.Index(q, "&"), strings.LastIndex(q, "&"))

	p, err = NewQueryParams("foo", func() {
		// this does not yield a string
	})
	require.NoError(t, err)
	_, err = p.GetQuery()
	require.Error(t, err)
	require.Equal(t, `unknown value type`, err.Error())
}

func TestQueryParams_Get(t *testing.T) {
	p, err := NewQueryParams()
	require.NoError(t, err)
	_, ok := p.Get("foo")
	require.False(t, ok)
	p.Add("foo", nil)
	v, ok := p.Get("foo")
	require.True(t, ok)
	require.Nil(t, v)
}

func TestQueryParams_GetIndex(t *testing.T) {
	p, err := NewQueryParams("foo", 1, "foo", 2, "foo", 3)
	require.NoError(t, err)
	v, ok := p.GetIndex("foo", 0)
	require.True(t, ok)
	require.Equal(t, 1, v)
	v, ok = p.GetIndex("foo", 1)
	require.True(t, ok)
	require.Equal(t, 2, v)
	v, ok = p.GetIndex("foo", 2)
	require.True(t, ok)
	require.Equal(t, 3, v)
	_, ok = p.GetIndex("foo", 3)
	require.False(t, ok)
	v, ok = p.GetIndex("foo", -1)
	require.True(t, ok)
	require.Equal(t, 3, v)
	v, ok = p.GetIndex("foo", -2)
	require.True(t, ok)
	require.Equal(t, 2, v)
	v, ok = p.GetIndex("foo", -3)
	require.True(t, ok)
	require.Equal(t, 1, v)
	_, ok = p.GetIndex("foo", -4)
	require.False(t, ok)
}

func TestQueryParams_Set(t *testing.T) {
	p, err := NewQueryParams("foo", 1, "foo", 2)
	require.NoError(t, err)
	v, ok := p.GetIndex("foo", 0)
	require.True(t, ok)
	require.Equal(t, 1, v)
	_, ok = p.GetIndex("foo", 1)
	require.True(t, ok)
	p.Set("foo", 3)
	v, ok = p.Get("foo")
	require.True(t, ok)
	require.Equal(t, 3, v)
	_, ok = p.GetIndex("foo", 1)
	require.False(t, ok)
}

func TestQueryParams_Add(t *testing.T) {
	p, err := NewQueryParams("foo", 1)
	require.NoError(t, err)
	v, ok := p.GetIndex("foo", 0)
	require.True(t, ok)
	require.Equal(t, 1, v)
	_, ok = p.GetIndex("foo", 1)
	require.False(t, ok)
	p.Add("foo", 2)
	v, ok = p.GetIndex("foo", 1)
	require.True(t, ok)
	require.Equal(t, 2, v)
}

func TestQueryParams_Del(t *testing.T) {
	p, err := NewQueryParams("foo", 1, "foo", 2)
	require.NoError(t, err)
	v, ok := p.GetIndex("foo", 1)
	require.True(t, ok)
	require.Equal(t, 2, v)
	p.Del("foo")
	_, ok = p.Get("foo")
	require.False(t, ok)
}

func TestQueryParams_Has(t *testing.T) {
	p, err := NewQueryParams("foo", 1, "foo", 2)
	require.NoError(t, err)
	require.True(t, p.Has("foo"))
	p.Del("foo")
	require.False(t, p.Has("foo"))
}

func TestQueryParams_Sorted(t *testing.T) {
	p, err := NewQueryParams("foo", 1, "baz", 2, "bar", 3)
	require.NoError(t, err)
	rp, ok := p.(*queryParams)
	require.True(t, ok)
	require.True(t, rp.sorted)
	q, err := p.GetQuery()
	require.NoError(t, err)
	require.Equal(t, `?bar=3&baz=2&foo=1`, q)
	p.Sorted(false)
	q, err = p.GetQuery()
	require.NoError(t, err)
	require.Contains(t, q, `foo=1`)
	require.Contains(t, q, `baz=2`)
	require.Contains(t, q, `bar=3`)
}

func TestQueryParams_Clone(t *testing.T) {
	p1, err := NewQueryParams("foo", 1)
	require.NoError(t, err)
	rp1, ok := p1.(*queryParams)
	require.True(t, ok)
	require.True(t, rp1.sorted)
	p2 := p1.Clone()
	rp2, ok := p2.(*queryParams)
	require.True(t, ok)
	require.True(t, rp2.sorted)
	p2.Sorted(false)
	require.False(t, rp2.sorted)
	require.True(t, rp1.sorted)

	_, ok = p1.Get("foo")
	require.True(t, ok)
	_, ok = p2.Get("foo")
	require.True(t, ok)
	p2.Del("foo")
	_, ok = p1.Get("foo")
	require.True(t, ok)
	_, ok = p2.Get("foo")
	require.False(t, ok)
}
