package e2e_test

import (
	"context"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := getCollectorBinary(t)

	t.Run("should shutdown gracefully on SIGINT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		configPath := getTestdataPath(t, "minimal.yaml")
		collectorCmd := exec.CommandContext(ctx, binary, "start", "--config", configPath)
		err := collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		err = collectorCmd.Process.Signal(syscall.SIGINT)
		require.NoError(t, err)

		err = collectorCmd.Wait()
		assert.NoError(t, err)
	})

	t.Run("should shutdown gracefully on SIGTERM", func(t *testing.T) {
		configPath := getTestdataPath(t, "minimal.yaml")
		collectorCmd := exec.Command(binary, "start", "--config", configPath)
		err := collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		err = collectorCmd.Process.Signal(syscall.SIGTERM)
		require.NoError(t, err)

		err = collectorCmd.Wait()
		assert.NoError(t, err)
	})
}
