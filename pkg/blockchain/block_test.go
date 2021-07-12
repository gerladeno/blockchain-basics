package blockchain

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBlockchain(t *testing.T) {
	started := time.Now()
	_, err := NewBlockchain()
	require.NoError(t, err)
	fmt.Printf("time elapsed in ns: %d", time.Since(started).Nanoseconds())
}
