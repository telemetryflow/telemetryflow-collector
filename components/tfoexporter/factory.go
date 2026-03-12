// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tfoexporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// TypeStr is the type string identifier for the TFO exporter.
	TypeStr = "tfo"

	// DefaultEndpoint is the default TFO Platform endpoint.
	DefaultEndpoint = "https://api.telemetryflow.id"
)

// NewFactory creates a new factory for the TFO exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(TypeStr),
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, component.StabilityLevelStable),
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelStable),
		exporter.WithLogs(createLogsExporter, component.StabilityLevelStable),
	)
}

// createDefaultConfig creates the default configuration for the exporter.
func createDefaultConfig() component.Config {
	return &Config{
		ClientConfig: confighttp.ClientConfig{
			Endpoint: DefaultEndpoint,
			Timeout:  30 * time.Second,
		},
		UseV2API: true,
		RetryConfig: configretry.BackOffConfig{
			Enabled:             true,
			InitialInterval:     5 * time.Second,
			MaxInterval:         30 * time.Second,
			MaxElapsedTime:      5 * time.Minute,
			RandomizationFactor: 0.5,
			Multiplier:          1.5,
		},
	}
}

// createTracesExporter creates a traces exporter.
func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	oCfg := cfg.(*Config)
	exp, err := newTFOExporter(oCfg, &set)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTraces(
		ctx,
		set,
		cfg,
		exp.pushTraces,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(oCfg.RetryConfig),
	)
}

// createMetricsExporter creates a metrics exporter.
func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	oCfg := cfg.(*Config)
	exp, err := newTFOExporter(oCfg, &set)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		exp.pushMetrics,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(oCfg.RetryConfig),
	)
}

// createLogsExporter creates a logs exporter.
func createLogsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	oCfg := cfg.(*Config)
	exp, err := newTFOExporter(oCfg, &set)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewLogs(
		ctx,
		set,
		cfg,
		exp.pushLogs,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(oCfg.RetryConfig),
	)
}
