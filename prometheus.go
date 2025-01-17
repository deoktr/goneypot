package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PrometheusAddr = "localhost"
	PrometheusPort = "9001"

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
			Help: "Goneypot config infos.",
		},
		[]string{"addr", "port", "private_key_file", "logging_root", "server_version", "prompt", "banner", "user", "password", "allow_login", "prom_addr", "prom_port", "tags"},
	)
	loginAtempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goneypot_login_atempts",
		Help: "Total login atempts.",
	})
	openedConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "goneypot_opened_connections",
		Help: "Current opened connections.",
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
		Help: "Total goneypot local errors.",
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
		"port":             Port,
		"private_key_file": PrivateKeyFile,
		"logging_root":     LoggingRoot,
		"server_version":   ServerVersion,
		"prompt":           Prompt,
		"banner":           Banner,
		"user":             User,
		"password":         Password,
		"allow_login":      fmt.Sprint(!DisableLogin),
		"prom_addr":        PrometheusAddr,
		"prom_port":        PrometheusPort,
		"tags":             "stringlabels",
	}).Inc()

	http.Handle("/metrics", promhttp.Handler())

	addr := PrometheusAddr + ":" + PrometheusPort

	log.Printf("starting prometheus metrics on: %s", addr)

	err := http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Print("metrics server closed")
	} else if err != nil {
		log.Printf("error starting metrics server: %s", err)
	}
}
