package urit

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetValueIf(t *testing.T) {
	dt := time.Date(2022, 11, 8, 12, 13, 14, 0, time.UTC)
	testCases := []struct {
		value     interface{}
		expectOk  bool
		expectStr string
	}{
		{
			"foo",
			true,
			"foo",
		},
		{
			1,
			true,
			"1",
		},
		{
			1.23,
			true,
			"1.23",
		},
		{
			true,
			true,
			"true",
		},
		{
			false,
			true,
			"false",
		},
		{
			dt,
			true,
			"2022-11-08T12:13:14Z",
		},
		{
			&dt,
			true,
			"2022-11-08T12:13:14Z",
		},
		{
			func() string {
				return "foo"
			},
			true,
			"foo",
		},
		{
			&valueStruct{Value: "foo"},
			true,
			"foo",
		},
		{
			&marshallable{Value: "foo"},
			true,
			"foo",
		},
		{
			&marshallable2{Value: true},
			true,
			"true",
		},
		{
			&marshallable2{Value: false},
			false,
			"",
		},
		{
			nil,
			false,
			"",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			str, ok := getValueIf(tc.value)
			if tc.expectOk {
				require.True(t, ok)
				require.Equal(t, tc.expectStr, str)
			} else {
				require.False(t, ok)
			}
		})
	}
}

type valueStruct struct {
	Value string
}

func (v *valueStruct) String() string {
	return v.Value
}

type marshallable struct {
	Value string
}

func (m *marshallable) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Value + `"`), nil
}

type marshallable2 struct {
	Value bool
}

func (m *marshallable2) MarshalJSON() ([]byte, error) {
	if m.Value {
		return []byte("true"), nil
	}
	return nil, errors.New("whoops")
}

func TestGetValue(t *testing.T) {
	_, err := getValue(nil)
	require.Error(t, err)
	require.Equal(t, `unknown value type`, err.Error())

	_, err = getValue(func() {
		// this does not yield a string
	})
	require.Error(t, err)
	require.Equal(t, `unknown value type`, err.Error())

	str, err := getValue("foo")
	require.NoError(t, err)
	require.Equal(t, "foo", str)
}
