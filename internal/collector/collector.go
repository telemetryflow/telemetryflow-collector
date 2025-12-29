// Package collector provides the core collector lifecycle management.
//
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
package collector

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/config"
	"github.com/telemetryflow/telemetryflow-collector/internal/exporter/debug"
	"github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
	"github.com/telemetryflow/telemetryflow-collector/internal/receiver/otlp"
)

// Collector is the main telemetry collector
type Collector struct {
	id     string
	config *config.Config
	logger *zap.Logger

	// Components
	otlpReceiver  *otlp.Receiver
	pipeline      *pipeline.Pipeline
	debugExporter *debug.Exporter

	// Health server
	healthServer *http.Server

	// State
	mu      sync.RWMutex
	running bool
	started time.Time
}

// New creates a new collector instance
func New(cfg *config.Config, logger *zap.Logger) (*Collector, error) {
	// Generate collector ID if not provided
	collectorID := cfg.Collector.ID
	if collectorID == "" {
		collectorID = uuid.New().String()
		logger.Info("Generated new collector ID", zap.String("id", collectorID))
	}

	c := &Collector{
		id:     collectorID,
		config: cfg,
		logger: logger,
	}

	// Initialize pipeline
	c.pipeline = pipeline.New(logger.Named("pipeline"))

	// Initialize debug exporter if verbosity is configured
	if cfg.Exporters.Debug.Verbosity != "" {
		c.debugExporter = debug.New(debug.Config{
			Verbosity: cfg.Exporters.Debug.Verbosity,
		}, logger.Named("debug-exporter"))

		// Add debug exporter to pipeline
		c.pipeline.AddTraceExporter(c.debugExporter)
		c.pipeline.AddMetricsExporter(c.debugExporter)
		c.pipeline.AddLogsExporter(c.debugExporter)

		logger.Info("Debug exporter enabled", zap.String("verbosity", cfg.Exporters.Debug.Verbosity))
	}

	// Initialize OTLP receiver if enabled
	if cfg.Receivers.OTLP.Enabled {
		otlpCfg := otlp.Config{
			GRPCEnabled:              cfg.Receivers.OTLP.Protocols.GRPC.Enabled,
			GRPCEndpoint:             cfg.Receivers.OTLP.Protocols.GRPC.Endpoint,
			GRPCMaxRecvMsgSizeMiB:    cfg.Receivers.OTLP.Protocols.GRPC.MaxRecvMsgSizeMiB,
			GRPCMaxConcurrentStreams: cfg.Receivers.OTLP.Protocols.GRPC.MaxConcurrentStreams,
			HTTPEnabled:              cfg.Receivers.OTLP.Protocols.HTTP.Enabled,
			HTTPEndpoint:             cfg.Receivers.OTLP.Protocols.HTTP.Endpoint,
		}

		c.otlpReceiver = otlp.New(otlpCfg, c.pipeline, logger.Named("otlp-receiver"))
		logger.Info("OTLP receiver configured",
			zap.Bool("grpc_enabled", otlpCfg.GRPCEnabled),
			zap.String("grpc_endpoint", otlpCfg.GRPCEndpoint),
			zap.Bool("http_enabled", otlpCfg.HTTPEnabled),
			zap.String("http_endpoint", otlpCfg.HTTPEndpoint),
		)
	}

	return c, nil
}

// ID returns the collector ID
func (c *Collector) ID() string {
	return c.id
}

// Run starts the collector and blocks until context is cancelled
func (c *Collector) Run(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("collector is already running")
	}
	c.running = true
	c.started = time.Now()
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.running = false
		c.mu.Unlock()
	}()

	c.logger.Info("Collector starting",
		zap.String("id", c.id),
		zap.String("hostname", c.config.Collector.Hostname),
		zap.String("name", c.config.Collector.Name),
	)

	// Start OTLP receiver
	if c.otlpReceiver != nil {
		if err := c.otlpReceiver.Start(ctx); err != nil {
			return fmt.Errorf("failed to start OTLP receiver: %w", err)
		}
	}

	// Start health check server if enabled
	if c.config.Extensions.Health.Enabled {
		go c.startHealthServer(ctx)
	}

	c.logger.Info("Collector started successfully",
		zap.String("id", c.id),
	)

	// Wait for context cancellation
	<-ctx.Done()
	c.logger.Info("Collector shutdown requested")
	return c.shutdown(context.Background())
}

// startHealthServer starts the health check server
func (c *Collector) startHealthServer(ctx context.Context) {
	healthCfg := c.config.Extensions.Health

	mux := http.NewServeMux()
	mux.HandleFunc(healthCfg.Path, func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		running := c.running
		c.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if running {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy"}`))
		}
	})

	// Add stats endpoint
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := c.Stats()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{
  "id": "%s",
  "hostname": "%s",
  "running": %t,
  "uptime_seconds": %.0f,
  "receiver": {
    "traces_received": %d,
    "metrics_received": %d,
    "logs_received": %d
  },
  "pipeline": {
    "traces_processed": %d,
    "metrics_processed": %d,
    "logs_processed": %d
  }
}`,
			stats.ID,
			stats.Hostname,
			stats.Running,
			stats.Uptime.Seconds(),
			stats.ReceiverStats.TracesReceived,
			stats.ReceiverStats.MetricsReceived,
			stats.ReceiverStats.LogsReceived,
			stats.PipelineStats.TracesProcessed,
			stats.PipelineStats.MetricsProcessed,
			stats.PipelineStats.LogsProcessed,
		)
	})

	c.healthServer = &http.Server{
		Addr:              healthCfg.Endpoint,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	c.logger.Info("Health check server listening",
		zap.String("endpoint", healthCfg.Endpoint),
		zap.String("path", healthCfg.Path),
	)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.healthServer.Shutdown(shutdownCtx)
	}()

	if err := c.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.logger.Error("Health check server error", zap.Error(err))
	}
}

// shutdown gracefully stops all components
func (c *Collector) shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down collector components")

	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	// Stop OTLP receiver
	if c.otlpReceiver != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.otlpReceiver.Stop(ctx); err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("OTLP receiver shutdown: %w", err))
				errMu.Unlock()
			}
		}()
	}

	// Stop health server
	if c.healthServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := c.healthServer.Shutdown(shutdownCtx); err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("health server shutdown: %w", err))
				errMu.Unlock()
			}
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("All components stopped")
	case <-time.After(15 * time.Second):
		c.logger.Warn("Shutdown timeout, some components may not have stopped cleanly")
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	uptime := time.Since(c.started)
	c.logger.Info("Collector shutdown complete", zap.Duration("uptime", uptime))
	return nil
}

// IsRunning returns whether the collector is running
func (c *Collector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Uptime returns the collector uptime
func (c *Collector) Uptime() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.running {
		return 0
	}
	return time.Since(c.started)
}

// Stats returns collector statistics
func (c *Collector) Stats() CollectorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CollectorStats{
		ID:       c.id,
		Hostname: c.config.Collector.Hostname,
		Running:  c.running,
		Started:  c.started,
		Uptime:   time.Since(c.started),
	}

	if c.otlpReceiver != nil {
		receiverStats := c.otlpReceiver.Stats()
		stats.ReceiverStats = ReceiverStats{
			TracesReceived:  receiverStats.TracesReceived,
			MetricsReceived: receiverStats.MetricsReceived,
			LogsReceived:    receiverStats.LogsReceived,
		}
	}

	if c.pipeline != nil {
		pipelineStats := c.pipeline.Stats()
		stats.PipelineStats = PipelineStats{
			TracesProcessed:  pipelineStats.TracesProcessed,
			MetricsProcessed: pipelineStats.MetricsProcessed,
			LogsProcessed:    pipelineStats.LogsProcessed,
		}
	}

	return stats
}

// CollectorStats contains collector statistics
type CollectorStats struct {
	ID            string        `json:"id"`
	Hostname      string        `json:"hostname"`
	Running       bool          `json:"running"`
	Started       time.Time     `json:"started"`
	Uptime        time.Duration `json:"uptime"`
	ReceiverStats ReceiverStats `json:"receiver"`
	PipelineStats PipelineStats `json:"pipeline"`
}

// ReceiverStats contains receiver statistics
type ReceiverStats struct {
	TracesReceived  int64 `json:"traces_received"`
	MetricsReceived int64 `json:"metrics_received"`
	LogsReceived    int64 `json:"logs_received"`
}

// PipelineStats contains pipeline statistics
type PipelineStats struct {
	TracesProcessed  int64 `json:"traces_processed"`
	MetricsProcessed int64 `json:"metrics_processed"`
	LogsProcessed    int64 `json:"logs_processed"`
}
