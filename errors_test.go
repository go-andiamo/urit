package urit

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemplateParseError_Error(t *testing.T) {
	var tpe error = newTemplateParseError("fooey", 16, nil)
	require.Equal(t, "fooey", tpe.Error())
}

func TestTemplateParseError_Unwrap(t *testing.T) {
	tpe := newTemplateParseError("fooey", 16, nil)
	require.Nil(t, tpe.Unwrap())
	require.Nil(t, errors.Unwrap(tpe))

	tpe = newTemplateParseError("fooey", 16, errors.New("wrapped fooey"))
	require.NotNil(t, tpe.Unwrap())
	require.NotNil(t, errors.Unwrap(tpe))
	require.Equal(t, "wrapped fooey", errors.Unwrap(tpe).Error())
}

func TestTemplateParseError_Position(t *testing.T) {
	tpe := newTemplateParseError("fooey", 16, nil)
	require.Equal(t, 16, tpe.Position())
}
