package e2e_test

import (
	"context"
	"os"
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

	t.Run("should shutdown gracefully on SIGINT", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "tfo-collector-test", "./cmd/tfo-collector")
		err := cmd.Run()
		require.NoError(t, err)
		defer os.Remove("tfo-collector-test")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "./tfo-collector-test", "start", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		err = collectorCmd.Process.Signal(syscall.SIGINT)
		require.NoError(t, err)

		err = collectorCmd.Wait()
		assert.NoError(t, err)
	})

	t.Run("should shutdown gracefully on SIGTERM", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "tfo-collector-test", "./cmd/tfo-collector")
		err := cmd.Run()
		require.NoError(t, err)
		defer os.Remove("tfo-collector-test")

		collectorCmd := exec.Command("./tfo-collector-test", "start", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		err = collectorCmd.Process.Signal(syscall.SIGTERM)
		require.NoError(t, err)

		err = collectorCmd.Wait()
		assert.NoError(t, err)
	})
}
