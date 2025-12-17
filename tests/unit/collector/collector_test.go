package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

func TestNewCollector(t *testing.T) {
	t.Run("should create collector with valid config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.NotEmpty(t, c.ID())
	})

	t.Run("should generate ID if not provided", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Collector.ID = ""
		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, c.ID())
	})
}