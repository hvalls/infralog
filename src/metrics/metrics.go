package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// ChangesTotal counts drift events by type and resource type
	ChangesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infralog_changes_total",
			Help: "Total number of infrastructure changes detected",
		},
		[]string{"type", "resource_type"},
	)

	// PollErrorsTotal counts polling errors by stage
	PollErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infralog_poll_errors_total",
			Help: "Total number of polling errors",
		},
		[]string{"stage"},
	)

	// LastSuccessfulPollTimestamp records when the last successful poll occurred
	LastSuccessfulPollTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "infralog_last_successful_poll_timestamp",
			Help: "Unix timestamp of the last successful poll",
		},
	)

	// NotificationsSentTotal counts successful notifications by target
	NotificationsSentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infralog_notifications_sent_total",
			Help: "Total number of notifications sent successfully",
		},
		[]string{"target"},
	)

	// NotificationErrorsTotal counts notification failures by target
	NotificationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infralog_notification_errors_total",
			Help: "Total number of notification errors",
		},
		[]string{"target"},
	)
)

func init() {
	prometheus.MustRegister(ChangesTotal)
	prometheus.MustRegister(PollErrorsTotal)
	prometheus.MustRegister(LastSuccessfulPollTimestamp)
	prometheus.MustRegister(NotificationsSentTotal)
	prometheus.MustRegister(NotificationErrorsTotal)
}

// Server wraps the HTTP server for metrics endpoint
type Server struct {
	server *http.Server
}

// NewServer creates a new metrics server
func NewServer(address string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		server: &http.Server{
			Addr:    address,
			Handler: mux,
		},
	}
}

// Start begins serving metrics in a goroutine
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()
	return nil
}

// Shutdown gracefully stops the metrics server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// RecordPollSuccess updates the last successful poll timestamp
func RecordPollSuccess() {
	LastSuccessfulPollTimestamp.SetToCurrentTime()
}

// RecordPollError increments the poll error counter for a given stage
func RecordPollError(stage string) {
	PollErrorsTotal.WithLabelValues(stage).Inc()
}

// RecordChange increments the change counter for a given type and resource type
func RecordChange(changeType, resourceType string) {
	ChangesTotal.WithLabelValues(changeType, resourceType).Inc()
}

// RecordNotificationSuccess increments the notification success counter
func RecordNotificationSuccess(target string) {
	NotificationsSentTotal.WithLabelValues(target).Inc()
}

// RecordNotificationError increments the notification error counter
func RecordNotificationError(target string) {
	NotificationErrorsTotal.WithLabelValues(target).Inc()
}

