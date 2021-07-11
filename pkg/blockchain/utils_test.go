package blockchain

import (
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestIntToHex(t *testing.T) {
	for i := 0; i < 1e6; i++ {
		n := rand.Int63n(math.MaxInt64)
		a := intToHex(n)
		b := IntToHex(n)
		require.NotEqual(t, a, b, "err on %dth iteration: n = %d", i, n)
	}
}