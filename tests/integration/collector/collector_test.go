package collector_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

func TestCollectorLifecycle(t *testing.T) {
	t.Run("should start and stop gracefully", func(t *testing.T) {
		cfg := config.DefaultConfig()
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
