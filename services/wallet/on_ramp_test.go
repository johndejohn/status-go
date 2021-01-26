package wallet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCryptoOnRamps_Get(t *testing.T) {
	cors := CryptoOnRampManager{}
	require.Equal(t, 0, len(cors.ramps))

	rs, err := cors.Get()
	require.NoError(t, err)
	require.Greater(t, len(rs), 0)
}
