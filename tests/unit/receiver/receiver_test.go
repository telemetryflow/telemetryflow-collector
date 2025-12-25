package receiver_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/tests/mocks"
)

func TestMockReceiver(t *testing.T) {
	t.Run("should start and stop receiver", func(t *testing.T) {
		receiver := mocks.NewMockReceiver("test-receiver")

		ctx := context.Background()

		receiver.On("Start", ctx).Return(nil)
		receiver.On("Stop").Return(nil)

		err := receiver.Start(ctx)
		require.NoError(t, err)
		assert.True(t, receiver.IsRunning())

		err = receiver.Stop()
		assert.NoError(t, err)
		assert.False(t, receiver.IsRunning())

		receiver.AssertExpectations(t)
	})

	t.Run("should receive telemetry data", func(t *testing.T) {
		receiver := mocks.NewMockReceiver("test-receiver")

		ctx := context.Background()
		receiver.On("Start", ctx).Return(nil)
		receiver.On("Stop").Return(nil)

		err := receiver.Start(ctx)
		require.NoError(t, err)

		testData := mocks.MockOTLPMetrics()
		receiver.SimulateReceive(testData)

		select {
		case data := <-receiver.DataChannel():
			assert.Equal(t, "metrics", data.Type)
			assert.NotZero(t, data.ReceivedAt)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for data")
		}

		_ = receiver.Stop()
		receiver.AssertExpectations(t)
	})
}
