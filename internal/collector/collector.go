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
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

// Collector is the main telemetry collector
type Collector struct {
	id     string
	config *config.Config
	logger *zap.Logger

	// Servers
	grpcServer *grpc.Server
	httpServer *http.Server

	// State
	mu      sync.RWMutex
	running bool
	started time.Time

	// Metrics
	metricsReceived int64
	logsReceived    int64
	tracesReceived  int64
	metricsExported int64
	logsExported    int64
	tracesExported  int64
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

	// Initialize gRPC server if enabled
	if cfg.Receivers.OTLP.Enabled && cfg.Receivers.OTLP.Protocols.GRPC.Enabled {
		if err := c.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}

	// Initialize HTTP server if enabled
	if cfg.Receivers.OTLP.Enabled && cfg.Receivers.OTLP.Protocols.HTTP.Enabled {
		if err := c.initHTTPServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
		}
	}

	return c, nil
}

// initGRPCServer initializes the gRPC OTLP receiver
func (c *Collector) initGRPCServer() error {
	grpcCfg := c.config.Receivers.OTLP.Protocols.GRPC

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(grpcCfg.MaxRecvMsgSizeMiB * 1024 * 1024),
		grpc.MaxConcurrentStreams(grpcCfg.MaxConcurrentStreams),
	}

	// TODO: Add TLS configuration if enabled
	// TODO: Add keepalive configuration

	c.grpcServer = grpc.NewServer(opts...)

	// TODO: Register OTLP services
	// ptraceotlp.RegisterGRPCServer(c.grpcServer, c)
	// pmetricotlp.RegisterGRPCServer(c.grpcServer, c)
	// plogotlp.RegisterGRPCServer(c.grpcServer, c)

	c.logger.Info("gRPC server initialized",
		zap.String("endpoint", grpcCfg.Endpoint),
	)

	return nil
}

// initHTTPServer initializes the HTTP OTLP receiver
func (c *Collector) initHTTPServer() error {
	httpCfg := c.config.Receivers.OTLP.Protocols.HTTP

	mux := http.NewServeMux()

	// OTLP endpoints
	mux.HandleFunc("/v1/metrics", c.handleMetrics)
	mux.HandleFunc("/v1/logs", c.handleLogs)
	mux.HandleFunc("/v1/traces", c.handleTraces)

	// Health endpoint
	mux.HandleFunc("/health", c.handleHealth)

	c.httpServer = &http.Server{
		Addr:           httpCfg.Endpoint,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	c.logger.Info("HTTP server initialized",
		zap.String("endpoint", httpCfg.Endpoint),
	)

	return nil
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
	)

	// Create error channel for component errors
	errChan := make(chan error, 2)

	// Start gRPC server if configured
	if c.grpcServer != nil {
		go func() {
			grpcEndpoint := c.config.Receivers.OTLP.Protocols.GRPC.Endpoint
			lis, err := net.Listen("tcp", grpcEndpoint)
			if err != nil {
				errChan <- fmt.Errorf("failed to listen on %s: %w", grpcEndpoint, err)
				return
			}
			c.logger.Info("gRPC server listening",
				zap.String("endpoint", grpcEndpoint),
			)
			if err := c.grpcServer.Serve(lis); err != nil {
				errChan <- fmt.Errorf("gRPC server error: %w", err)
			}
		}()
	}

	// Start HTTP server if configured
	if c.httpServer != nil {
		go func() {
			c.logger.Info("HTTP server listening",
				zap.String("endpoint", c.httpServer.Addr),
			)
			if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()
	}

	// Start health check server if enabled
	if c.config.Extensions.Health.Enabled {
		go c.startHealthServer(ctx)
	}

	c.logger.Info("Collector started successfully")

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		c.logger.Info("Collector shutdown requested")
		return c.shutdown()
	case err := <-errChan:
		c.logger.Error("Component error, initiating shutdown", zap.Error(err))
		return err
	}
}

// startHealthServer starts the health check server
func (c *Collector) startHealthServer(ctx context.Context) {
	healthCfg := c.config.Extensions.Health

	mux := http.NewServeMux()
	mux.HandleFunc(healthCfg.Path, func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		running := c.running
		c.mu.RUnlock()

		if running {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy"}`))
		}
	})

	server := &http.Server{
		Addr:    healthCfg.Endpoint,
		Handler: mux,
	}

	c.logger.Info("Health check server listening",
		zap.String("endpoint", healthCfg.Endpoint),
		zap.String("path", healthCfg.Path),
	)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.logger.Error("Health check server error", zap.Error(err))
	}
}

// shutdown gracefully stops all components
func (c *Collector) shutdown() error {
	c.logger.Info("Shutting down collector components")

	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	// Stop gRPC server
	if c.grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.grpcServer.GracefulStop()
			c.logger.Info("gRPC server stopped")
		}()
	}

	// Stop HTTP server
	if c.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := c.httpServer.Shutdown(shutdownCtx); err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("HTTP server shutdown: %w", err))
				errMu.Unlock()
			} else {
				c.logger.Info("HTTP server stopped")
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

// HTTP Handlers

// handleMetrics handles OTLP metrics requests
func (c *Collector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Parse and process OTLP metrics
	c.mu.Lock()
	c.metricsReceived++
	c.mu.Unlock()

	c.logger.Debug("Received metrics", zap.String("content_type", r.Header.Get("Content-Type")))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// handleLogs handles OTLP logs requests
func (c *Collector) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Parse and process OTLP logs
	c.mu.Lock()
	c.logsReceived++
	c.mu.Unlock()

	c.logger.Debug("Received logs", zap.String("content_type", r.Header.Get("Content-Type")))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// handleTraces handles OTLP traces requests
func (c *Collector) handleTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Parse and process OTLP traces
	c.mu.Lock()
	c.tracesReceived++
	c.mu.Unlock()

	c.logger.Debug("Received traces", zap.String("content_type", r.Header.Get("Content-Type")))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// handleHealth handles health check requests
func (c *Collector) handleHealth(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()

	if running {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unhealthy"}`))
	}
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

	return CollectorStats{
		ID:              c.id,
		Hostname:        c.config.Collector.Hostname,
		Running:         c.running,
		Started:         c.started,
		Uptime:          time.Since(c.started),
		MetricsReceived: c.metricsReceived,
		LogsReceived:    c.logsReceived,
		TracesReceived:  c.tracesReceived,
		MetricsExported: c.metricsExported,
		LogsExported:    c.logsExported,
		TracesExported:  c.tracesExported,
	}
}

// CollectorStats contains collector statistics
type CollectorStats struct {
	ID              string        `json:"id"`
	Hostname        string        `json:"hostname"`
	Running         bool          `json:"running"`
	Started         time.Time     `json:"started"`
	Uptime          time.Duration `json:"uptime"`
	MetricsReceived int64         `json:"metrics_received"`
	LogsReceived    int64         `json:"logs_received"`
	TracesReceived  int64         `json:"traces_received"`
	MetricsExported int64         `json:"metrics_exported"`
	LogsExported    int64         `json:"logs_exported"`
	TracesExported  int64         `json:"traces_exported"`
}
