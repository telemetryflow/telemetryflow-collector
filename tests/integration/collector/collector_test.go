package collector_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

// getFreePort returns a free port by letting OS assign one
func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = listener.Close() }()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func TestCollectorLifecycle(t *testing.T) {
	t.Run("should start and stop gracefully", func(t *testing.T) {
		cfg := config.DefaultConfig()

		// Use dynamic ports to avoid conflicts
		grpcPort, err := getFreePort()
		require.NoError(t, err)
		httpPort, err := getFreePort()
		require.NoError(t, err)

		cfg.Receivers.OTLP.Protocols.GRPC.Endpoint = fmt.Sprintf("0.0.0.0:%d", grpcPort)
		cfg.Receivers.OTLP.Protocols.HTTP.Endpoint = fmt.Sprintf("0.0.0.0:%d", httpPort)

		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- c.Run(ctx)
		}()

		time.Sleep(100 * time.Millisecond)
		assert.True(t, c.IsRunning())

		cancel()
		err = <-errChan
		assert.NoError(t, err)
		assert.False(t, c.IsRunning())
	})

	t.Run("should return collector stats", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)

		stats := c.Stats()
		assert.NotEmpty(t, stats.ID)
		assert.False(t, stats.Running)
	})
}
