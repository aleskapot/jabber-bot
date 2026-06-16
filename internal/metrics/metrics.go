package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	MessagesSent       *prometheus.CounterVec
	APIDuration        *prometheus.HistogramVec
	FileUploadSize     prometheus.Histogram
	WebhookDeliveries  *prometheus.CounterVec
	WebhookDuration    *prometheus.HistogramVec
	WebhookRetries     *prometheus.CounterVec
	WebhookQueueLength prometheus.Gauge
}

var (
	registry   *prometheus.Registry
	appMetrics *Metrics
)

func init() {
	registry = prometheus.NewRegistry()
	appMetrics = newMetrics()
}

// newMetrics creates and registers all Prometheus metrics
func newMetrics() *Metrics {
	m := &Metrics{
		MessagesSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jabber_messages_sent_total",
				Help: "Total number of XMPP messages sent",
			},
			[]string{"to", "type"},
		),
		APIDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "jabber_api_request_duration_seconds",
				Help:    "Duration of API requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status_code"},
		),
		FileUploadSize: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "jabber_file_upload_size_bytes",
				Help:    "Size of uploaded files in bytes",
				Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 52428800, 104857600},
			},
		),
		WebhookDeliveries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jabber_webhook_deliveries_total",
				Help: "Total number of webhook delivery attempts",
			},
			[]string{"status"},
		),
		WebhookDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "jabber_webhook_duration_seconds",
				Help:    "Duration of webhook delivery attempts in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"status"},
		),
		WebhookRetries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jabber_webhook_retries_total",
				Help: "Total number of webhook retries",
			},
			[]string{"status"},
		),
		WebhookQueueLength: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "jabber_webhook_queue_length",
				Help: "Current length of the webhook message queue",
			},
		),
	}

	// Register all metrics with the custom registry
	registry.MustRegister(
		m.MessagesSent,
		m.APIDuration,
		m.FileUploadSize,
		m.WebhookDeliveries,
		m.WebhookDuration,
		m.WebhookRetries,
		m.WebhookQueueLength,
	)

	return m
}

// Get returns the singleton metrics instance
func Get() *Metrics {
	return appMetrics
}

// Handler returns an HTTP handler for the /metrics endpoint
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// RecordMessageSent increments the messages sent counter
func RecordMessageSent(to, msgType string) {
	appMetrics.MessagesSent.WithLabelValues(to, msgType).Inc()
}

// RecordAPIDuration observes the duration of an API request
func RecordAPIDuration(method, path, statusCode string, duration float64) {
	appMetrics.APIDuration.WithLabelValues(method, path, statusCode).Observe(duration)
}

// RecordFileUploadSize observes the size of an uploaded file
func RecordFileUploadSize(sizeBytes float64) {
	appMetrics.FileUploadSize.Observe(sizeBytes)
}

// RecordWebhookDelivery increments the webhook deliveries counter
func RecordWebhookDelivery(status string) {
	appMetrics.WebhookDeliveries.WithLabelValues(status).Inc()
}

// RecordWebhookDuration observes the duration of a webhook delivery
func RecordWebhookDuration(status string, duration float64) {
	appMetrics.WebhookDuration.WithLabelValues(status).Observe(duration)
}

// RecordWebhookRetry increments the webhook retries counter
func RecordWebhookRetry(status string) {
	appMetrics.WebhookRetries.WithLabelValues(status).Inc()
}

// SetWebhookQueueLength sets the current webhook queue length
func SetWebhookQueueLength(length int) {
	appMetrics.WebhookQueueLength.Set(float64(length))
}
