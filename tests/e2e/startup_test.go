package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	"os/exec"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := getCollectorBinary(t)

	t.Run("should start with default config", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		configPath := getTestdataPath(t, "minimal.yaml")
		collectorCmd := exec.CommandContext(ctx, binary, "start", "--config", configPath)
		err := collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)
		_ = collectorCmd.Process.Signal(os.Interrupt)
		err = collectorCmd.Wait()
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid config", func(t *testing.T) {
		configPath := getTestdataPath(t, "invalid.yaml")
		collectorCmd := exec.Command(binary, "start", "--config", configPath)
		err := collectorCmd.Run()
		assert.Error(t, err)
	})
}
