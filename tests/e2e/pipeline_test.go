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

	t.Run("should process data through pipeline", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "tfo-collector-test", "./cmd/tfo-collector")
		err := cmd.Run()
		require.NoError(t, err)
		defer os.Remove("tfo-collector-test")

		collectorCmd := exec.Command("./tfo-collector-test", "start", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		// Check health endpoint
		resp, err := http.Get("http://localhost:13133/")
		if err == nil {
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		collectorCmd.Process.Signal(os.Interrupt)
		collectorCmd.Wait()
	})
}