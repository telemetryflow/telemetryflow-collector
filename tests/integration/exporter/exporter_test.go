package exporter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/tests/mocks"
)

func TestExporterIntegration(t *testing.T) {
	t.Run("should export data successfully", func(t *testing.T) {
		exporter := mocks.NewMockExporter("test-exporter")

		ctx := context.Background()
		testData := []byte(`{"test": "data"}`)

		exporter.On("Start", ctx).Return(nil)
		exporter.On("Export", ctx, testData).Return(nil)
		exporter.On("Stop").Return(nil)

		err := exporter.Start(ctx)
		require.NoError(t, err)

		err = exporter.Export(ctx, testData)
		assert.NoError(t, err)

		err = exporter.Stop()
		assert.NoError(t, err)

		exporter.AssertExpectations(t)
	})
}
