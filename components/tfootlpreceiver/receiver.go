// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfootlpreceiver

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// tfoOTLPReceiver is the TFO-enhanced OTLP receiver with v1/v2 endpoint support.
type tfoOTLPReceiver struct {
	cfg      *Config
	settings *receiver.Settings
	logger   *zap.Logger

	// Consumers
	tracesConsumer  consumer.Traces
	metricsConsumer consumer.Metrics
	logsConsumer    consumer.Logs

	// Servers
	grpcServer *grpc.Server
	httpServer *http.Server

	// State
	mu      sync.RWMutex
	started bool

	// Metrics
	tracesReceived  atomic.Int64
	metricsReceived atomic.Int64
	logsReceived    atomic.Int64

	// Shared instance management
	shutdownWG sync.WaitGroup
}

var (
	// receiverInstance is a shared receiver instance for all signal types.
	receiverInstance     *tfoOTLPReceiver
	receiverInstanceLock sync.Mutex
)

// newTFOOTLPReceiver creates a new TFO OTLP receiver or returns an existing shared instance.
func newTFOOTLPReceiver(cfg *Config, set *receiver.Settings) (*tfoOTLPReceiver, error) {
	receiverInstanceLock.Lock()
	defer receiverInstanceLock.Unlock()

	if receiverInstance != nil {
		return receiverInstance, nil
	}

	r := &tfoOTLPReceiver{
		cfg:      cfg,
		settings: set,
		logger:   set.Logger,
	}

	receiverInstance = r
	return r, nil
}

// registerTracesConsumer registers a traces consumer.
func (r *tfoOTLPReceiver) registerTracesConsumer(tc consumer.Traces) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracesConsumer = tc
}

// registerMetricsConsumer registers a metrics consumer.
func (r *tfoOTLPReceiver) registerMetricsConsumer(mc consumer.Metrics) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metricsConsumer = mc
}

// registerLogsConsumer registers a logs consumer.
func (r *tfoOTLPReceiver) registerLogsConsumer(lc consumer.Logs) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logsConsumer = lc
}

// Start implements component.Component.
func (r *tfoOTLPReceiver) Start(ctx context.Context, host component.Host) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return nil
	}
	r.started = true
	r.mu.Unlock()

	// Start gRPC server if configured
	if r.cfg.Protocols.GRPC != nil {
		if err := r.startGRPC(ctx); err != nil {
			return err
		}
	}

	// Start HTTP server if configured
	if r.cfg.Protocols.HTTP != nil {
		if err := r.startHTTP(ctx); err != nil {
			return err
		}
	}

	r.logger.Info("TFO OTLP receiver started",
		zap.Bool("grpc_enabled", r.cfg.Protocols.GRPC != nil),
		zap.Bool("http_enabled", r.cfg.Protocols.HTTP != nil),
		zap.Bool("v2_endpoints", r.cfg.EnableV2Endpoints),
	)

	return nil
}

// startGRPC starts the gRPC server.
func (r *tfoOTLPReceiver) startGRPC(ctx context.Context) error {
	endpoint := r.cfg.Protocols.GRPC.NetAddr.Endpoint
	if endpoint == "" {
		endpoint = DefaultGRPCEndpoint
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(4 * 1024 * 1024), // 4 MiB default
	}

	r.grpcServer = grpc.NewServer(opts...)

	// Register OTLP gRPC services
	ptraceotlp.RegisterGRPCServer(r.grpcServer, &traceServer{r: r})
	pmetricotlp.RegisterGRPCServer(r.grpcServer, &metricsServer{r: r})
	plogotlp.RegisterGRPCServer(r.grpcServer, &logsServer{r: r})

	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		return err
	}

	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()
		r.logger.Info("TFO OTLP gRPC server listening", zap.String("endpoint", endpoint))
		if err := r.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			r.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// startHTTP starts the HTTP server with v1 and v2 endpoints.
func (r *tfoOTLPReceiver) startHTTP(ctx context.Context) error {
	endpoint := r.cfg.Protocols.HTTP.Endpoint
	if endpoint == "" {
		endpoint = DefaultHTTPEndpoint
	}

	mux := http.NewServeMux()

	// v1 endpoints (OTEL standard)
	tracesPath := r.cfg.Protocols.HTTP.TracesURLPath
	if tracesPath == "" {
		tracesPath = defaultTracesURLPath
	}
	metricsPath := r.cfg.Protocols.HTTP.MetricsURLPath
	if metricsPath == "" {
		metricsPath = defaultMetricsURLPath
	}
	logsPath := r.cfg.Protocols.HTTP.LogsURLPath
	if logsPath == "" {
		logsPath = defaultLogsURLPath
	}

	mux.HandleFunc(tracesPath, r.handleTraces)
	mux.HandleFunc(metricsPath, r.handleMetrics)
	mux.HandleFunc(logsPath, r.handleLogs)

	r.logger.Info("TFO OTLP HTTP v1 endpoints registered",
		zap.String("traces", tracesPath),
		zap.String("metrics", metricsPath),
		zap.String("logs", logsPath),
	)

	// v2 endpoints (TFO Platform) - served on same port
	if r.cfg.EnableV2Endpoints {
		mux.HandleFunc("/v2/traces", r.handleTraces)
		mux.HandleFunc("/v2/metrics", r.handleMetrics)
		mux.HandleFunc("/v2/logs", r.handleLogs)

		r.logger.Info("TFO OTLP HTTP v2 endpoints registered",
			zap.String("traces", "/v2/traces"),
			zap.String("metrics", "/v2/metrics"),
			zap.String("logs", "/v2/logs"),
		)
	}

	r.httpServer = &http.Server{
		Addr:           endpoint,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()
		r.logger.Info("TFO OTLP HTTP server listening", zap.String("endpoint", endpoint))
		if err := r.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

// Shutdown implements component.Component.
func (r *tfoOTLPReceiver) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	if !r.started {
		r.mu.Unlock()
		return nil
	}
	r.started = false
	r.mu.Unlock()

	if r.grpcServer != nil {
		r.grpcServer.GracefulStop()
	}

	if r.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := r.httpServer.Shutdown(shutdownCtx); err != nil {
			r.logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	r.shutdownWG.Wait()

	// Clear shared instance
	receiverInstanceLock.Lock()
	receiverInstance = nil
	receiverInstanceLock.Unlock()

	r.logger.Info("TFO OTLP receiver stopped",
		zap.Int64("traces_received", r.tracesReceived.Load()),
		zap.Int64("metrics_received", r.metricsReceived.Load()),
		zap.Int64("logs_received", r.logsReceived.Load()),
	)

	return nil
}

// =============================================================================
// gRPC Service Implementations
// =============================================================================

type traceServer struct {
	ptraceotlp.UnimplementedGRPCServer
	r *tfoOTLPReceiver
}

func (s *traceServer) Export(ctx context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	td := req.Traces()
	spanCount := td.SpanCount()
	s.r.tracesReceived.Add(int64(spanCount))

	s.r.logger.Debug("Received traces via gRPC",
		zap.Int("span_count", spanCount),
		zap.Int("resource_spans", td.ResourceSpans().Len()),
	)

	if s.r.tracesConsumer != nil {
		if err := s.r.tracesConsumer.ConsumeTraces(ctx, td); err != nil {
			s.r.logger.Error("Failed to consume traces", zap.Error(err))
			return ptraceotlp.NewExportResponse(), err
		}
	}

	return ptraceotlp.NewExportResponse(), nil
}

type metricsServer struct {
	pmetricotlp.UnimplementedGRPCServer
	r *tfoOTLPReceiver
}

func (s *metricsServer) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	md := req.Metrics()
	dataPointCount := md.DataPointCount()
	s.r.metricsReceived.Add(int64(dataPointCount))

	s.r.logger.Debug("Received metrics via gRPC",
		zap.Int("data_point_count", dataPointCount),
		zap.Int("resource_metrics", md.ResourceMetrics().Len()),
	)

	if s.r.metricsConsumer != nil {
		if err := s.r.metricsConsumer.ConsumeMetrics(ctx, md); err != nil {
			s.r.logger.Error("Failed to consume metrics", zap.Error(err))
			return pmetricotlp.NewExportResponse(), err
		}
	}

	return pmetricotlp.NewExportResponse(), nil
}

type logsServer struct {
	plogotlp.UnimplementedGRPCServer
	r *tfoOTLPReceiver
}

func (s *logsServer) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	ld := req.Logs()
	logRecordCount := ld.LogRecordCount()
	s.r.logsReceived.Add(int64(logRecordCount))

	s.r.logger.Debug("Received logs via gRPC",
		zap.Int("log_record_count", logRecordCount),
		zap.Int("resource_logs", ld.ResourceLogs().Len()),
	)

	if s.r.logsConsumer != nil {
		if err := s.r.logsConsumer.ConsumeLogs(ctx, ld); err != nil {
			s.r.logger.Error("Failed to consume logs", zap.Error(err))
			return plogotlp.NewExportResponse(), err
		}
	}

	return plogotlp.NewExportResponse(), nil
}

// =============================================================================
// HTTP Handlers
// =============================================================================

// TFO Authentication Headers
const (
	headerKeyID       = "X-TelemetryFlow-Key-ID"
	headerKeySecret   = "X-TelemetryFlow-Key-Secret"
	headerCollectorID = "X-TelemetryFlow-Collector-ID"
)

// isV2Endpoint checks if the request path is a v2 endpoint.
func isV2Endpoint(path string) bool {
	return path == "/v2/traces" || path == "/v2/metrics" || path == "/v2/logs"
}

// validateV2Auth validates TFO authentication for v2 endpoints.
// Returns true if auth is valid, false otherwise.
func (r *tfoOTLPReceiver) validateV2Auth(w http.ResponseWriter, req *http.Request) bool {
	// Skip auth if not required
	if !r.cfg.V2Auth.Required {
		return true
	}

	keyID := req.Header.Get(headerKeyID)
	keySecret := req.Header.Get(headerKeySecret)

	// Check if API Key ID is present
	if keyID == "" {
		r.logger.Warn("v2 endpoint access denied: missing API Key ID",
			zap.String("path", req.URL.Path),
			zap.String("remote_addr", req.RemoteAddr),
		)
		http.Error(w, `{"error": "missing TelemetryFlow API Key ID"}`, http.StatusUnauthorized)
		return false
	}

	// Validate API Key ID format (should start with tfk_)
	if len(keyID) < 4 || keyID[:4] != "tfk_" {
		r.logger.Warn("v2 endpoint access denied: invalid API Key ID format",
			zap.String("path", req.URL.Path),
			zap.String("remote_addr", req.RemoteAddr),
		)
		http.Error(w, `{"error": "invalid TelemetryFlow API Key ID format (expected tfk_xxx)"}`, http.StatusUnauthorized)
		return false
	}

	// Check against valid API Key IDs if configured
	if len(r.cfg.V2Auth.ValidAPIKeyIDs) > 0 {
		valid := false
		for _, validID := range r.cfg.V2Auth.ValidAPIKeyIDs {
			if keyID == validID {
				valid = true
				break
			}
		}
		if !valid {
			r.logger.Warn("v2 endpoint access denied: API Key ID not in allowed list",
				zap.String("path", req.URL.Path),
				zap.String("key_id", keyID),
				zap.String("remote_addr", req.RemoteAddr),
			)
			http.Error(w, `{"error": "API Key ID not authorized"}`, http.StatusForbidden)
			return false
		}
	}

	// Validate secret if required
	if r.cfg.V2Auth.ValidateSecret {
		if keySecret == "" {
			r.logger.Warn("v2 endpoint access denied: missing API Key Secret",
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", req.RemoteAddr),
			)
			http.Error(w, `{"error": "missing TelemetryFlow API Key Secret"}`, http.StatusUnauthorized)
			return false
		}

		// Validate API Key Secret format (should start with tfs_)
		if len(keySecret) < 4 || keySecret[:4] != "tfs_" {
			r.logger.Warn("v2 endpoint access denied: invalid API Key Secret format",
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", req.RemoteAddr),
			)
			http.Error(w, `{"error": "invalid TelemetryFlow API Key Secret format (expected tfs_xxx)"}`, http.StatusUnauthorized)
			return false
		}
	}

	r.logger.Debug("v2 endpoint auth validated",
		zap.String("path", req.URL.Path),
		zap.String("key_id", keyID),
		zap.String("collector_id", req.Header.Get(headerCollectorID)),
	)

	return true
}

func (r *tfoOTLPReceiver) handleTraces(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate v2 authentication if this is a v2 endpoint
	if isV2Endpoint(req.URL.Path) && !r.validateV2Auth(w, req) {
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

	isV2 := isV2Endpoint(req.URL.Path)
	r.logger.Debug("Received traces via HTTP",
		zap.Int("span_count", spanCount),
		zap.String("path", req.URL.Path),
		zap.Bool("v2_endpoint", isV2),
	)

	if r.tracesConsumer != nil {
		if err := r.tracesConsumer.ConsumeTraces(req.Context(), td); err != nil {
			r.logger.Error("Failed to consume traces", zap.Error(err))
			http.Error(w, "Failed to process traces", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (r *tfoOTLPReceiver) handleMetrics(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate v2 authentication if this is a v2 endpoint
	if isV2Endpoint(req.URL.Path) && !r.validateV2Auth(w, req) {
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

	isV2 := isV2Endpoint(req.URL.Path)
	r.logger.Debug("Received metrics via HTTP",
		zap.Int("data_point_count", dataPointCount),
		zap.String("path", req.URL.Path),
		zap.Bool("v2_endpoint", isV2),
	)

	if r.metricsConsumer != nil {
		if err := r.metricsConsumer.ConsumeMetrics(req.Context(), md); err != nil {
			r.logger.Error("Failed to consume metrics", zap.Error(err))
			http.Error(w, "Failed to process metrics", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (r *tfoOTLPReceiver) handleLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate v2 authentication if this is a v2 endpoint
	if isV2Endpoint(req.URL.Path) && !r.validateV2Auth(w, req) {
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

	isV2 := isV2Endpoint(req.URL.Path)
	r.logger.Debug("Received logs via HTTP",
		zap.Int("log_record_count", logRecordCount),
		zap.String("path", req.URL.Path),
		zap.Bool("v2_endpoint", isV2),
	)

	if r.logsConsumer != nil {
		if err := r.logsConsumer.ConsumeLogs(req.Context(), ld); err != nil {
			r.logger.Error("Failed to consume logs", zap.Error(err))
			http.Error(w, "Failed to process logs", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}
