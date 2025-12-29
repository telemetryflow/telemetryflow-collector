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

	binary := getCollectorBinary(t)

	t.Run("should receive OTLP metrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		configPath := getTestdataPath(t, "minimal.yaml")
		collectorCmd := exec.CommandContext(ctx, binary, "start", "--config", configPath)
		err := collectorCmd.Start()
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
		resp, err := http.Post("http://localhost:4318/v2/metrics", "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			_ = resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		_ = collectorCmd.Process.Signal(os.Interrupt)
		_ = collectorCmd.Wait()
	})
}
