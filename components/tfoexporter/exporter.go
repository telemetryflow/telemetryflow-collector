// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoexporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
)

const (
	// TFO authentication headers
	headerKeyID       = "X-TelemetryFlow-Key-ID"
	headerKeySecret   = "X-TelemetryFlow-Key-Secret"
	headerCollectorID = "X-TelemetryFlow-Collector-ID"
)

// tfoExporter is the TFO Platform exporter with auto-auth injection.
type tfoExporter struct {
	cfg      *Config
	settings *exporter.Settings
	logger   *zap.Logger
	client   *http.Client

	// Auth credentials (resolved from config or extension)
	apiKeyID     string
	apiKeySecret string
	collectorID  string

	// Metrics
	tracesExported  atomic.Int64
	metricsExported atomic.Int64
	logsExported    atomic.Int64
}

// newTFOExporter creates a new TFO exporter.
func newTFOExporter(cfg *Config, set *exporter.Settings) (*tfoExporter, error) {
	return &tfoExporter{
		cfg:      cfg,
		settings: set,
		logger:   set.Logger,
	}, nil
}

// start initializes the exporter.
func (e *tfoExporter) start(ctx context.Context, host component.Host) error {
	// Create HTTP client using the new API
	httpClient, err := e.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), e.settings.TelemetrySettings)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	e.client = httpClient

	// Resolve authentication credentials
	if e.cfg.Auth != nil {
		if e.cfg.Auth.Extension.String() != "" {
			// Get credentials from tfoauth extension
			ext, ok := host.GetExtensions()[e.cfg.Auth.Extension]
			if !ok {
				return fmt.Errorf("tfoauth extension %q not found", e.cfg.Auth.Extension)
			}

			// Try to get credentials from extension
			if authProvider, ok := ext.(AuthProvider); ok {
				e.apiKeyID = authProvider.GetAPIKeyID()
				e.apiKeySecret = authProvider.GetAPIKeySecret()
			}
		} else {
			// Use direct credentials from config
			e.apiKeyID = string(e.cfg.Auth.APIKeyID)
			e.apiKeySecret = string(e.cfg.Auth.APIKeySecret)
		}
	}

	// Resolve collector identity
	if e.cfg.CollectorIdentity.String() != "" {
		ext, ok := host.GetExtensions()[e.cfg.CollectorIdentity]
		if !ok {
			e.logger.Warn("tfoidentity extension not found, collector ID will not be set",
				zap.String("extension", e.cfg.CollectorIdentity.String()),
			)
		} else {
			if identityProvider, ok := ext.(IdentityProvider); ok {
				e.collectorID = identityProvider.GetCollectorID()
			}
		}
	}

	e.logger.Info("TFO exporter started",
		zap.String("endpoint", e.cfg.Endpoint),
		zap.Bool("use_v2_api", e.cfg.UseV2API),
		zap.Bool("has_auth", e.apiKeyID != ""),
		zap.Bool("has_collector_id", e.collectorID != ""),
	)

	return nil
}

// shutdown stops the exporter.
func (e *tfoExporter) shutdown(ctx context.Context) error {
	e.logger.Info("TFO exporter stopped",
		zap.Int64("traces_exported", e.tracesExported.Load()),
		zap.Int64("metrics_exported", e.metricsExported.Load()),
		zap.Int64("logs_exported", e.logsExported.Load()),
	)
	return nil
}

// pushTraces exports traces to the TFO Platform.
func (e *tfoExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	req := ptraceotlp.NewExportRequestFromTraces(td)
	data, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal traces: %w", err)
	}

	endpoint := e.cfg.Endpoint + e.cfg.GetTracesEndpoint()
	if err := e.sendData(ctx, endpoint, data, "application/x-protobuf"); err != nil {
		return err
	}

	e.tracesExported.Add(int64(td.SpanCount()))
	e.logger.Debug("Exported traces",
		zap.Int("span_count", td.SpanCount()),
		zap.String("endpoint", endpoint),
	)

	return nil
}

// pushMetrics exports metrics to the TFO Platform.
func (e *tfoExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	req := pmetricotlp.NewExportRequestFromMetrics(md)
	data, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	endpoint := e.cfg.Endpoint + e.cfg.GetMetricsEndpoint()
	if err := e.sendData(ctx, endpoint, data, "application/x-protobuf"); err != nil {
		return err
	}

	e.metricsExported.Add(int64(md.DataPointCount()))
	e.logger.Debug("Exported metrics",
		zap.Int("data_point_count", md.DataPointCount()),
		zap.String("endpoint", endpoint),
	)

	return nil
}

// pushLogs exports logs to the TFO Platform.
func (e *tfoExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	req := plogotlp.NewExportRequestFromLogs(ld)
	data, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	endpoint := e.cfg.Endpoint + e.cfg.GetLogsEndpoint()
	if err := e.sendData(ctx, endpoint, data, "application/x-protobuf"); err != nil {
		return err
	}

	e.logsExported.Add(int64(ld.LogRecordCount()))
	e.logger.Debug("Exported logs",
		zap.Int("log_record_count", ld.LogRecordCount()),
		zap.String("endpoint", endpoint),
	)

	return nil
}

// sendData sends data to the TFO Platform with authentication headers.
func (e *tfoExporter) sendData(ctx context.Context, endpoint string, data []byte, contentType string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	// Inject TFO authentication headers
	if e.apiKeyID != "" {
		req.Header.Set(headerKeyID, e.apiKeyID)
	}
	if e.apiKeySecret != "" {
		req.Header.Set(headerKeySecret, e.apiKeySecret)
	}
	if e.collectorID != "" {
		req.Header.Set(headerCollectorID, e.collectorID)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and discard body
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// AuthProvider is an interface for extensions that provide TFO authentication.
type AuthProvider interface {
	GetAPIKeyID() string
	GetAPIKeySecret() string
}

// IdentityProvider is an interface for extensions that provide collector identity.
type IdentityProvider interface {
	GetCollectorID() string
}
