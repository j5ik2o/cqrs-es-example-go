package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveArgsUsesHandlerWhenNoSubcommand(t *testing.T) {
	require.Equal(t, []string{"recoverRmuWorker"}, resolveArgs(nil, "recoverRmuWorker"))
}

func TestResolveArgsKeepsExplicitSubcommand(t *testing.T) {
	require.Equal(t, []string{"rmu"}, resolveArgs([]string{"rmu"}, "recoverRmuWorker"))
}

func TestResolveArgsIgnoresEmptyHandler(t *testing.T) {
	require.Empty(t, resolveArgs(nil, ""))
}
