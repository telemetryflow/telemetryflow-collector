package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// ProcessRequest represents a processing request
type ProcessRequest struct {
	Data       []byte            `json:"data"`
	DataType   string            `json:"data_type"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ProcessResponse represents a processing response
type ProcessResponse struct {
	Data         []byte `json:"data"`
	ItemsDropped int    `json:"items_dropped"`
	Error        string `json:"error,omitempty"`
}

// MockProcessor is a mock implementation of the Processor interface
type MockProcessor struct {
	mock.Mock
	name    string
	running bool
}

// NewMockProcessor creates a new mock processor
func NewMockProcessor(name string) *MockProcessor {
	return &MockProcessor{
		name: name,
	}
}

// Name returns the processor name
func (m *MockProcessor) Name() string {
	return m.name
}

// Start mocks starting the processor
func (m *MockProcessor) Start(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = true
	return args.Error(0)
}

// Stop mocks stopping the processor
func (m *MockProcessor) Stop() error {
	args := m.Called()
	m.running = false
	return args.Error(0)
}

// Process mocks processing data
func (m *MockProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

// IsRunning returns whether the processor is running
func (m *MockProcessor) IsRunning() bool {
	return m.running
}