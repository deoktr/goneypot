package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PrometheusAddr = "localhost"
	PrometheusPort = 9001

	buildInfo = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goneypot_build_info",
			Help: "Goneypot build infos.",
		},
		[]string{"version", "goarch", "goos", "goversion", "tags"},
	)
	configInfo = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goneypot_config_info",
			Help: "Goneypot configuration.",
		},
		[]string{"addr", "port", "private_key_file", "logging_root", "server_version", "prompt", "banner", "creds_file", "allow_login", "prom_addr", "prom_port", "tags"},
	)
	loginAtempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_login_atempts",
		Help: "Total login atempts.",
	})
	loginFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_login_failed",
		Help: "Total login failed.",
	})
	openedConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "goneypot_opened_connections",
		Help: "Currently opened connections.",
	})
	totalConnections = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_total_connections",
		Help: "Total opened connections.",
	})
	totalCommands = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_total_commands",
		Help: "Total commands sent.",
	})
	totalErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_errors",
		Help: "Total goneypot errors.",
	})
	requestDurations = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "goneypot_connection_duration_seconds",
		Help: "A histogram of the goneypot connection durations in seconds.",
		// from 0.1s to 1000s (16min)
		Buckets: prometheus.ExponentialBuckets(0.1, 10, 5),
	})
)

func startPrometheusListener() {
	buildInfo.With(prometheus.Labels{
		"version":   VERSION,
		"goarch":    runtime.GOARCH,
		"goos":      runtime.GOOS,
		"goversion": runtime.Version(),
		"tags":      "stringlabels",
	}).Inc()

	configInfo.With(prometheus.Labels{
		"addr":             Addr,
		"port":             fmt.Sprint(Port),
		"private_key_file": PrivateKeyFile,
		"logging_root":     LoggingRoot,
		"server_version":   ServerVersion,
		"prompt":           Prompt,
		"banner":           Banner,
		"creds_file":       CredsFile,
		"allow_login":      fmt.Sprint(!DisableLogin),
		"prom_addr":        PrometheusAddr,
		"prom_port":        fmt.Sprint(PrometheusPort),
		"tags":             "stringlabels",
	}).Inc()

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf("%s:%d", PrometheusAddr, PrometheusPort)

	logger.Printf("starting prometheus metrics on: %s", addr)

	err := http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Print("metrics server closed")
	} else if err != nil {
		logger.Printf("error starting metrics server: %s", err)
	}
}
