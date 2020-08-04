//go:generate  enumer -type=MetricType  -json -text -sql

package monitor

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricType int

const (
	CounterVec MetricType = iota + 1
	Counter
	GaugeVec
	Gauge
	HistogramVec
	Histogram
	SummaryVec
	Summary
)

var MetricNameError = errors.New("metric name CAN NOT be empty")
var MetricTypeError = errors.New("metric got wrong type")

// NewMetric associates monitor.Collector based on Metric.Type
func NewMetric(m *Metric) (prometheus.Collector, error) {
	var metric prometheus.Collector
	if m.Name == "" {
		return metric, MetricNameError
	}
	switch m.Type {
	case CounterVec:
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
			m.Labels,
		)
	case Counter:
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
		)
	case GaugeVec:
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
			m.Labels,
		)
	case Gauge:
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
		)
	case HistogramVec:
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
			m.Labels,
		)
	case Histogram:
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
		)
	case SummaryVec:
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
			m.Labels,
		)
	case Summary:
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Name:      m.Name,
				Help:      m.Help,
			},
		)
	default:
		return metric, MetricTypeError
	}
	return metric, nil
}

type Metric struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Type      MetricType
	Labels    []string
}

type Prometheus struct {
	reqCnt       *prometheus.CounterVec
	reqDur       *prometheus.HistogramVec
	reqSz, resSz prometheus.Summary

	MetricsPath string
	Namespace   string
	Subsystem   string
}

func NewPrometheus(namespace string, subsystem string, metricsPath string) *Prometheus {
	p := &Prometheus{
		Namespace:   namespace,
		Subsystem:   subsystem,
		MetricsPath: metricsPath,
	}
	p.registerMetrics()
	return p
}

func (p *Prometheus) registerMetrics() {
	reqCnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "handler", "host", "url"},
	)
	p.reqCnt = reqCnt
	prometheus.MustRegister(reqCnt)

	reqDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "request_duration_millisecond",
			Help:      "The HTTP request latencies in Millisecond.",
		},
		[]string{"code", "method", "url"},
	)
	p.reqDur = reqDur
	prometheus.MustRegister(reqDur)
	resSz := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)
	p.resSz = resSz
	prometheus.MustRegister(resSz)
	reqSz := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: p.Namespace,
			Subsystem: p.Subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)
	p.reqSz = reqSz
	prometheus.MustRegister(reqSz)
}

func (p *Prometheus) prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (p *Prometheus) Use(s gin.IRoutes) {
	s.GET(p.MetricsPath, p.prometheusHandler())
	s.Use(p.HandlerFunc())

}

func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := c.Request.ContentLength

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Millisecond)
		resSz := float64(c.Writer.Size())

		p.reqDur.WithLabelValues(status, c.Request.Method, path).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, path).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}
}
