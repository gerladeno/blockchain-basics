package blockchain

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateWallet(t *testing.T) {
	w, err := CreateWallet()
	require.NoError(t, err)
	fmt.Printf("%s\n", w.PrivateKey)
	addr, err := w.GetAddress()
	require.NoError(t, err)
	fmt.Printf("%s\n", addr)
}
