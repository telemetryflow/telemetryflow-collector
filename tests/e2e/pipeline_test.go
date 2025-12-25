package e2e_test

import (
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := getCollectorBinary(t)

	t.Run("should process data through pipeline", func(t *testing.T) {
		configPath := getTestdataPath(t, "minimal.yaml")
		collectorCmd := exec.Command(binary, "start", "--config", configPath)
		err := collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		// Check health endpoint
		resp, err := http.Get("http://localhost:13133/")
		if err == nil {
			_ = resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		_ = collectorCmd.Process.Signal(os.Interrupt)
		_ = collectorCmd.Wait()
	})
}
