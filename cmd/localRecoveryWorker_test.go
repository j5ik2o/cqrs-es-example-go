package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeLocalRecoveryWorkerInterval(t *testing.T) {
	require.Equal(t, defaultLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(0))
	require.Equal(t, defaultLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(-time.Second))
	require.Equal(t, minLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(time.Millisecond))
	require.Equal(t, 2*time.Second,
		normalizeLocalRecoveryWorkerInterval(2*time.Second))
}
