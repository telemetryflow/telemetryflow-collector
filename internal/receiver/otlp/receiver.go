// Package otlp provides OTLP receiver implementation for the TelemetryFlow Collector.
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
package otlp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Consumer is the interface for consuming telemetry data
type Consumer interface {
	ConsumeTraces(ctx context.Context, td ptrace.Traces) error
	ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error
	ConsumeLogs(ctx context.Context, ld plog.Logs) error
}

// Config holds the OTLP receiver configuration
type Config struct {
	// gRPC configuration
	GRPCEnabled              bool
	GRPCEndpoint             string
	GRPCMaxRecvMsgSizeMiB    int
	GRPCMaxConcurrentStreams uint32

	// HTTP configuration
	HTTPEnabled  bool
	HTTPEndpoint string
}

// Receiver is the OTLP receiver that handles both gRPC and HTTP protocols
type Receiver struct {
	config   Config
	logger   *zap.Logger
	consumer Consumer

	// Servers
	grpcServer *grpc.Server
	httpServer *http.Server

	// State
	mu      sync.RWMutex
	running bool

	// Metrics
	tracesReceived  atomic.Int64
	metricsReceived atomic.Int64
	logsReceived    atomic.Int64
}

// New creates a new OTLP receiver
func New(cfg Config, consumer Consumer, logger *zap.Logger) *Receiver {
	return &Receiver{
		config:   cfg,
		logger:   logger,
		consumer: consumer,
	}
}

// Start starts the OTLP receiver
func (r *Receiver) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("receiver already running")
	}
	r.running = true
	r.mu.Unlock()

	errChan := make(chan error, 2)

	// Start gRPC server if enabled
	if r.config.GRPCEnabled {
		if err := r.startGRPC(ctx, errChan); err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}

	// Start HTTP server if enabled
	if r.config.HTTPEnabled {
		if err := r.startHTTP(ctx, errChan); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}

	r.logger.Info("OTLP receiver started",
		zap.Bool("grpc", r.config.GRPCEnabled),
		zap.String("grpc_endpoint", r.config.GRPCEndpoint),
		zap.Bool("http", r.config.HTTPEnabled),
		zap.String("http_endpoint", r.config.HTTPEndpoint),
	)

	return nil
}

// startGRPC starts the gRPC server
func (r *Receiver) startGRPC(ctx context.Context, errChan chan<- error) error {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(r.config.GRPCMaxRecvMsgSizeMiB * 1024 * 1024),
		grpc.MaxConcurrentStreams(r.config.GRPCMaxConcurrentStreams),
	}

	r.grpcServer = grpc.NewServer(opts...)

	// Register OTLP services using wrapper types
	ptraceotlp.RegisterGRPCServer(r.grpcServer, &traceServer{r: r})
	pmetricotlp.RegisterGRPCServer(r.grpcServer, &metricsServer{r: r})
	plogotlp.RegisterGRPCServer(r.grpcServer, &logsServer{r: r})

	lis, err := net.Listen("tcp", r.config.GRPCEndpoint)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", r.config.GRPCEndpoint, err)
	}

	go func() {
		r.logger.Info("OTLP gRPC server listening", zap.String("endpoint", r.config.GRPCEndpoint))
		if err := r.grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	return nil
}

// startHTTP starts the HTTP server
func (r *Receiver) startHTTP(ctx context.Context, errChan chan<- error) error {
	mux := http.NewServeMux()

	// OTLP HTTP endpoints
	// TelemetryFlow Platform v2 endpoints
	mux.HandleFunc("/v2/traces", r.handleTraces)
	mux.HandleFunc("/v2/metrics", r.handleMetrics)
	mux.HandleFunc("/v2/logs", r.handleLogs)

	// Legacy v1 endpoints for backwards compatibility
	mux.HandleFunc("/v1/traces", r.handleTraces)
	mux.HandleFunc("/v1/metrics", r.handleMetrics)
	mux.HandleFunc("/v1/logs", r.handleLogs)

	r.httpServer = &http.Server{
		Addr:           r.config.HTTPEndpoint,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		r.logger.Info("OTLP HTTP server listening", zap.String("endpoint", r.config.HTTPEndpoint))
		if err := r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	return nil
}

// Stop stops the OTLP receiver
func (r *Receiver) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = false
	r.mu.Unlock()

	var wg sync.WaitGroup

	if r.grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.grpcServer.GracefulStop()
			r.logger.Info("OTLP gRPC server stopped")
		}()
	}

	if r.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := r.httpServer.Shutdown(shutdownCtx); err != nil {
				r.logger.Error("HTTP server shutdown error", zap.Error(err))
			} else {
				r.logger.Info("OTLP HTTP server stopped")
			}
		}()
	}

	wg.Wait()
	return nil
}

// =============================================================================
// gRPC Service Implementations
// =============================================================================

// traceServer wraps the receiver to implement ptraceotlp.GRPCServer
type traceServer struct {
	ptraceotlp.UnimplementedGRPCServer
	r *Receiver
}

func (s *traceServer) Export(ctx context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	td := req.Traces()
	spanCount := td.SpanCount()

	s.r.tracesReceived.Add(int64(spanCount))

	s.r.logger.Debug("Received traces via gRPC",
		zap.Int("span_count", spanCount),
		zap.Int("resource_spans", td.ResourceSpans().Len()),
	)

	if s.r.consumer != nil {
		if err := s.r.consumer.ConsumeTraces(ctx, td); err != nil {
			s.r.logger.Error("Failed to consume traces", zap.Error(err))
			return ptraceotlp.NewExportResponse(), err
		}
	}

	return ptraceotlp.NewExportResponse(), nil
}

// metricsServer wraps the receiver to implement pmetricotlp.GRPCServer
type metricsServer struct {
	pmetricotlp.UnimplementedGRPCServer
	r *Receiver
}

func (s *metricsServer) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	md := req.Metrics()
	dataPointCount := md.DataPointCount()

	s.r.metricsReceived.Add(int64(dataPointCount))

	s.r.logger.Debug("Received metrics via gRPC",
		zap.Int("data_point_count", dataPointCount),
		zap.Int("resource_metrics", md.ResourceMetrics().Len()),
	)

	if s.r.consumer != nil {
		if err := s.r.consumer.ConsumeMetrics(ctx, md); err != nil {
			s.r.logger.Error("Failed to consume metrics", zap.Error(err))
			return pmetricotlp.NewExportResponse(), err
		}
	}

	return pmetricotlp.NewExportResponse(), nil
}

// logsServer wraps the receiver to implement plogotlp.GRPCServer
type logsServer struct {
	plogotlp.UnimplementedGRPCServer
	r *Receiver
}

func (s *logsServer) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	ld := req.Logs()
	logRecordCount := ld.LogRecordCount()

	s.r.logsReceived.Add(int64(logRecordCount))

	s.r.logger.Debug("Received logs via gRPC",
		zap.Int("log_record_count", logRecordCount),
		zap.Int("resource_logs", ld.ResourceLogs().Len()),
	)

	if s.r.consumer != nil {
		if err := s.r.consumer.ConsumeLogs(ctx, ld); err != nil {
			s.r.logger.Error("Failed to consume logs", zap.Error(err))
			return plogotlp.NewExportResponse(), err
		}
	}

	return plogotlp.NewExportResponse(), nil
}

// =============================================================================
// HTTP Handlers
// =============================================================================

// handleTraces handles HTTP OTLP trace requests
func (r *Receiver) handleTraces(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = req.Body.Close() }()

	// Determine content type and unmarshal accordingly
	contentType := req.Header.Get("Content-Type")
	exportReq := ptraceotlp.NewExportRequest()

	var unmarshalErr error
	if contentType == "application/json" {
		unmarshalErr = exportReq.UnmarshalJSON(body)
	} else {
		unmarshalErr = exportReq.UnmarshalProto(body)
	}

	if unmarshalErr != nil {
		r.logger.Error("Failed to unmarshal traces", zap.Error(unmarshalErr), zap.String("content_type", contentType))
		http.Error(w, "Failed to unmarshal traces", http.StatusBadRequest)
		return
	}

	td := exportReq.Traces()
	spanCount := td.SpanCount()
	r.tracesReceived.Add(int64(spanCount))

	r.logger.Debug("Received traces via HTTP",
		zap.Int("span_count", spanCount),
		zap.Int("resource_spans", td.ResourceSpans().Len()),
	)

	if r.consumer != nil {
		if err := r.consumer.ConsumeTraces(req.Context(), td); err != nil {
			r.logger.Error("Failed to consume traces", zap.Error(err))
			http.Error(w, "Failed to process traces", http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// handleMetrics handles HTTP OTLP metrics requests
func (r *Receiver) handleMetrics(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = req.Body.Close() }()

	contentType := req.Header.Get("Content-Type")
	exportReq := pmetricotlp.NewExportRequest()

	var unmarshalErr error
	if contentType == "application/json" {
		unmarshalErr = exportReq.UnmarshalJSON(body)
	} else {
		unmarshalErr = exportReq.UnmarshalProto(body)
	}

	if unmarshalErr != nil {
		r.logger.Error("Failed to unmarshal metrics", zap.Error(unmarshalErr), zap.String("content_type", contentType))
		http.Error(w, "Failed to unmarshal metrics", http.StatusBadRequest)
		return
	}

	md := exportReq.Metrics()
	dataPointCount := md.DataPointCount()
	r.metricsReceived.Add(int64(dataPointCount))

	r.logger.Debug("Received metrics via HTTP",
		zap.Int("data_point_count", dataPointCount),
		zap.Int("resource_metrics", md.ResourceMetrics().Len()),
	)

	if r.consumer != nil {
		if err := r.consumer.ConsumeMetrics(req.Context(), md); err != nil {
			r.logger.Error("Failed to consume metrics", zap.Error(err))
			http.Error(w, "Failed to process metrics", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// handleLogs handles HTTP OTLP logs requests
func (r *Receiver) handleLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = req.Body.Close() }()

	contentType := req.Header.Get("Content-Type")
	exportReq := plogotlp.NewExportRequest()

	var unmarshalErr error
	if contentType == "application/json" {
		unmarshalErr = exportReq.UnmarshalJSON(body)
	} else {
		unmarshalErr = exportReq.UnmarshalProto(body)
	}

	if unmarshalErr != nil {
		r.logger.Error("Failed to unmarshal logs", zap.Error(unmarshalErr), zap.String("content_type", contentType))
		http.Error(w, "Failed to unmarshal logs", http.StatusBadRequest)
		return
	}

	ld := exportReq.Logs()
	logRecordCount := ld.LogRecordCount()
	r.logsReceived.Add(int64(logRecordCount))

	r.logger.Debug("Received logs via HTTP",
		zap.Int("log_record_count", logRecordCount),
		zap.Int("resource_logs", ld.ResourceLogs().Len()),
	)

	if r.consumer != nil {
		if err := r.consumer.ConsumeLogs(req.Context(), ld); err != nil {
			r.logger.Error("Failed to consume logs", zap.Error(err))
			http.Error(w, "Failed to process logs", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// Stats returns receiver statistics
func (r *Receiver) Stats() ReceiverStats {
	return ReceiverStats{
		TracesReceived:  r.tracesReceived.Load(),
		MetricsReceived: r.metricsReceived.Load(),
		LogsReceived:    r.logsReceived.Load(),
	}
}

// ReceiverStats contains receiver statistics
type ReceiverStats struct {
	TracesReceived  int64 `json:"traces_received"`
	MetricsReceived int64 `json:"metrics_received"`
	LogsReceived    int64 `json:"logs_received"`
}
