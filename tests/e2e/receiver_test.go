package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTLPReceiver(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("should receive OTLP metrics", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "tfo-collector-test", "./cmd/tfo-collector")
		err := cmd.Run()
		require.NoError(t, err)
		defer os.Remove("tfo-collector-test")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "./tfo-collector-test", "start", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		time.Sleep(3 * time.Second)

		// Send test metrics
		metricsData := map[string]interface{}{
			"resourceMetrics": []map[string]interface{}{
				{
					"resource": map[string]interface{}{
						"attributes": []map[string]interface{}{
							{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
						},
					},
					"scopeMetrics": []map[string]interface{}{
						{
							"metrics": []map[string]interface{}{
								{
									"name": "test.metric",
									"sum": map[string]interface{}{
										"dataPoints": []map[string]interface{}{
											{"asInt": "100", "timeUnixNano": time.Now().UnixNano()},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		jsonData, _ := json.Marshal(metricsData)
		resp, err := http.Post("http://localhost:4318/v1/metrics", "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		collectorCmd.Process.Signal(os.Interrupt)
		collectorCmd.Wait()
	})
}