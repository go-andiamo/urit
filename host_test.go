package urit

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewHost(t *testing.T) {
	h := NewHost(`www.example.com`)
	require.NotNil(t, h)
	rh, ok := h.(*host)
	require.True(t, ok)
	require.Equal(t, `www.example.com`, rh.address)
	require.Equal(t, `www.example.com`, h.GetAddress())
}
