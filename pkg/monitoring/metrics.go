package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// Memory operations
	MemoryOperations *prometheus.CounterVec
	MemoryLatency    *prometheus.HistogramVec

	// LLM operations
	LLMRequests *prometheus.CounterVec
	LLMLatency  *prometheus.HistogramVec
	LLMTokens   *prometheus.CounterVec

	// Vector operations
	VectorSearches *prometheus.CounterVec
	VectorLatency  *prometheus.HistogramVec

	// Evolution operations
	EvolutionRuns    *prometheus.CounterVec
	EvolutionLatency *prometheus.HistogramVec

	// System metrics
	ActiveConnections prometheus.Gauge
	ErrorRate         *prometheus.CounterVec

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
}

// MetricsServer manages the metrics HTTP server
type MetricsServer struct {
	server  *http.Server
	metrics *Metrics
	logger  *zap.Logger
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		MemoryOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_memory_operations_total",
				Help: "Total number of memory operations",
			},
			[]string{"operation", "status"},
		),
		MemoryLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zetmem_memory_operation_duration_seconds",
				Help:    "Memory operation latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		LLMRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_llm_requests_total",
				Help: "Total number of LLM requests",
			},
			[]string{"model", "operation", "status"},
		),
		LLMLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zetmem_llm_request_duration_seconds",
				Help:    "LLM request latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"model", "operation"},
		),
		LLMTokens: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_llm_tokens_total",
				Help: "Total number of LLM tokens used",
			},
			[]string{"model", "type"}, // type: prompt, completion
		),
		VectorSearches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_vector_searches_total",
				Help: "Total number of vector searches",
			},
			[]string{"status"},
		),
		VectorLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zetmem_vector_search_duration_seconds",
				Help:    "Vector search latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{},
		),
		EvolutionRuns: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_evolution_runs_total",
				Help: "Total number of evolution runs",
			},
			[]string{"trigger_type", "status"},
		),
		EvolutionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zetmem_evolution_duration_seconds",
				Help:    "Evolution process latency",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300}, // Evolution can take longer
			},
			[]string{"trigger_type"},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "zetmem_active_connections",
				Help: "Number of active MCP connections",
			},
		),
		ErrorRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_errors_total",
				Help: "Total number of errors",
			},
			[]string{"component", "error_type"},
		),
		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zetmem_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		metrics.MemoryOperations,
		metrics.MemoryLatency,
		metrics.LLMRequests,
		metrics.LLMLatency,
		metrics.LLMTokens,
		metrics.VectorSearches,
		metrics.VectorLatency,
		metrics.EvolutionRuns,
		metrics.EvolutionLatency,
		metrics.ActiveConnections,
		metrics.ErrorRate,
		metrics.CacheHits,
		metrics.CacheMisses,
	)

	return metrics
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int, logger *zap.Logger) *MetricsServer {
	metrics := NewMetrics()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &MetricsServer{
		server:  server,
		metrics: metrics,
		logger:  logger,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start(ctx context.Context) error {
	ms.logger.Info("Starting metrics server", zap.String("addr", ms.server.Addr))

	go func() {
		<-ctx.Done()
		ms.logger.Info("Shutting down metrics server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ms.server.Shutdown(shutdownCtx)
	}()

	if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// GetMetrics returns the metrics instance
func (ms *MetricsServer) GetMetrics() *Metrics {
	return ms.metrics
}

// Timer is a helper for timing operations
type Timer struct {
	start time.Time
	hist  prometheus.Observer
}

// NewTimer creates a new timer
func NewTimer(hist prometheus.Observer) *Timer {
	return &Timer{
		start: time.Now(),
		hist:  hist,
	}
}

// Observe records the elapsed time
func (t *Timer) Observe() {
	t.hist.Observe(time.Since(t.start).Seconds())
}

// ObserveDuration records a specific duration
func (t *Timer) ObserveDuration(duration time.Duration) {
	t.hist.Observe(duration.Seconds())
}

// Middleware for HTTP request metrics
func (ms *MetricsServer) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active connections
		ms.metrics.ActiveConnections.Inc()
		defer ms.metrics.ActiveConnections.Dec()

		// Call next handler
		next.ServeHTTP(w, r)

		// Record latency
		duration := time.Since(start)
		ms.logger.Debug("HTTP request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration))
	})
}

// Helper methods for common metric operations

// RecordMemoryOperation records a memory operation
func (m *Metrics) RecordMemoryOperation(operation, status string, duration time.Duration) {
	m.MemoryOperations.WithLabelValues(operation, status).Inc()
	m.MemoryLatency.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordLLMRequest records an LLM request
func (m *Metrics) RecordLLMRequest(model, operation, status string, duration time.Duration, promptTokens, completionTokens int) {
	m.LLMRequests.WithLabelValues(model, operation, status).Inc()
	m.LLMLatency.WithLabelValues(model, operation).Observe(duration.Seconds())
	m.LLMTokens.WithLabelValues(model, "prompt").Add(float64(promptTokens))
	m.LLMTokens.WithLabelValues(model, "completion").Add(float64(completionTokens))
}

// RecordVectorSearch records a vector search operation
func (m *Metrics) RecordVectorSearch(status string, duration time.Duration) {
	m.VectorSearches.WithLabelValues(status).Inc()
	m.VectorLatency.WithLabelValues().Observe(duration.Seconds())
}

// RecordEvolution records an evolution operation
func (m *Metrics) RecordEvolution(triggerType, status string, duration time.Duration) {
	m.EvolutionRuns.WithLabelValues(triggerType, status).Inc()
	m.EvolutionLatency.WithLabelValues(triggerType).Observe(duration.Seconds())
}

// RecordError records an error
func (m *Metrics) RecordError(component, errorType string) {
	m.ErrorRate.WithLabelValues(component, errorType).Inc()
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(cacheType string) {
	m.CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(cacheType string) {
	m.CacheMisses.WithLabelValues(cacheType).Inc()
}
