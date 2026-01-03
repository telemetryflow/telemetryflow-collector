// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfootlpreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	// TypeStr is the type string identifier for the TFO OTLP receiver.
	TypeStr = "tfootlp"

	// DefaultGRPCEndpoint is the default gRPC endpoint.
	DefaultGRPCEndpoint = "0.0.0.0:4317"

	// DefaultHTTPEndpoint is the default HTTP endpoint.
	DefaultHTTPEndpoint = "0.0.0.0:4318"

	// Default URL paths for OTLP v1 (standard)
	defaultTracesURLPath  = "/v1/traces"
	defaultMetricsURLPath = "/v1/metrics"
	defaultLogsURLPath    = "/v1/logs"
)

// NewFactory creates a new factory for the TFO OTLP receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelStable),
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelStable),
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelStable),
	)
}

// createDefaultConfig creates the default configuration for the receiver.
//
// Default behavior:
//   - v1 endpoints (/v1/traces, /v1/metrics, /v1/logs): Open, no auth required (Community)
//   - v2 endpoints (/v2/traces, /v2/metrics, /v2/logs): Requires TFO auth headers (Platform)
func createDefaultConfig() component.Config {
	return &Config{
		Protocols: ProtocolsConfig{
			GRPC: &GRPCConfig{},
			HTTP: &HTTPConfig{
				TracesURLPath:  defaultTracesURLPath,
				MetricsURLPath: defaultMetricsURLPath,
				LogsURLPath:    defaultLogsURLPath,
			},
		},
		EnableV2Endpoints: true,
		V2Auth: V2AuthConfig{
			Required:       true,  // v2 endpoints require TFO auth by default
			ValidateSecret: false, // Only validate API Key ID presence by default
		},
	}
}

// createTracesReceiver creates a traces receiver.
func createTracesReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (receiver.Traces, error) {
	oCfg := cfg.(*Config)
	r, err := newTFOOTLPReceiver(oCfg, &set)
	if err != nil {
		return nil, err
	}
	r.registerTracesConsumer(nextConsumer)
	return r, nil
}

// createMetricsReceiver creates a metrics receiver.
func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	oCfg := cfg.(*Config)
	r, err := newTFOOTLPReceiver(oCfg, &set)
	if err != nil {
		return nil, err
	}
	r.registerMetricsConsumer(nextConsumer)
	return r, nil
}

// createLogsReceiver creates a logs receiver.
func createLogsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (receiver.Logs, error) {
	oCfg := cfg.(*Config)
	r, err := newTFOOTLPReceiver(oCfg, &set)
	if err != nil {
		return nil, err
	}
	r.registerLogsConsumer(nextConsumer)
	return r, nil
}
