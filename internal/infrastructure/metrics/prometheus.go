package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Prometheus struct {
	httpRequestsTotal           *prometheus.CounterVec
	httpRequestsDurationSeconds *prometheus.HistogramVec
	tgRequestsTotal             *prometheus.CounterVec
	tgRequestsDurationSeconds   *prometheus.HistogramVec
	activeLinksTotal            *prometheus.GaugeVec
	scrapeDurationSeconds       *prometheus.HistogramVec
}

func NewPrometheus(name string) *Prometheus {
	httpRequestsTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestsDurationSeconds := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name + "_http_requests_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	tgRequestsTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_tg_requests_total",
			Help: "Total number of TG requests",
		},
		[]string{"state", "status"},
	)
	tgRequestsDurationSeconds := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name + "_tg_requests_duration_seconds",
			Help:    "Histogram of TG request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"state"},
	)
	activeLinksTotal := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name + "_active_links_total",
		},
		[]string{"type"},
	)
	scrapeDurationSeconds := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name + "_scrape_duration_seconds",
			Help:    "Histogram of scrape durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	return &Prometheus{
		httpRequestsTotal:           httpRequestsTotal,
		httpRequestsDurationSeconds: httpRequestsDurationSeconds,
		tgRequestsTotal:             tgRequestsTotal,
		tgRequestsDurationSeconds:   tgRequestsDurationSeconds,
		activeLinksTotal:            activeLinksTotal,
		scrapeDurationSeconds:       scrapeDurationSeconds,
	}
}

func (p *Prometheus) IncHTTPRequestsTotal(method, path string, status int) {
	p.httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%dxx", status/100)).Inc()
}

func (p *Prometheus) ObserveHTTPRequestsDurationSeconds(method, path string, seconds float64) {
	p.httpRequestsDurationSeconds.WithLabelValues(method, path).Observe(seconds)
}

func (p *Prometheus) IncTGRequestsTotal(state, status string) {
	p.tgRequestsTotal.WithLabelValues(state, status).Inc()
}

func (p *Prometheus) ObserveTGRequestsDurationSeconds(state string, seconds float64) {
	p.tgRequestsDurationSeconds.WithLabelValues(state).Observe(seconds)
}

func (p *Prometheus) IncActiveLinksTotal(linkType string) {
	p.activeLinksTotal.WithLabelValues(linkType).Inc()
}

func (p *Prometheus) DecActiveLinksTotal(linkType string) {
	p.activeLinksTotal.WithLabelValues(linkType).Dec()
}

func (p *Prometheus) ObserveScrapeDurationSeconds(scrapeType string, seconds float64) {
	p.scrapeDurationSeconds.WithLabelValues(scrapeType).Observe(seconds)
}
