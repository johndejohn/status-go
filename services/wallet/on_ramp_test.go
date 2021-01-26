package wallet

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCryptoOnRamps_Get(t *testing.T) {
	cors := CryptoOnRamps{}
	require.Equal(t, 0, len(cors))

	err := cors.Get()
	require.NoError(t, err)
	require.Greater(t, len(cors), 0)
}
