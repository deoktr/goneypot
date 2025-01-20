package main

import (
	"flag"
	"fmt"
	"os"
)

func ExecuteCLI() {
	flag.Parse()

	startHoneypot()
}

func init() {
	// logging
	flag.StringVar(&LoggingRoot, "logroot", LoggingRoot, "logging root directory (default current)")
	flag.StringVar(&CredsLoggerFile, "creds-log-file", CredsLoggerFile, "login attemp credentials log file")
	flag.BoolVar(&DisableCredsLog, "disable-creds-log", DisableCredsLog, "disable credentials logging")

	// honeypot server
	flag.StringVar(&Addr, "addr", Addr, "honeypot listen address")
	flag.IntVar(&Port, "port", Port, "honeypot listen port")
	flag.StringVar(&PrivateKeyFile, "key", PrivateKeyFile, "private SSH key file")
	flag.StringVar(&ServerVersion, "server-version", ServerVersion, "ssh server version")
	flag.StringVar(&Banner, "banner", Banner, "ssh banner")

	// shell
	flag.StringVar(&Prompt, "prompt", Prompt, "shell prompt")

	// authentication
	flag.StringVar(&CredsFile, "creds-file", CredsFile, "file containing login credentials")
	flag.BoolVar(&DisableLogin, "disable-login", DisableLogin, "disable login")

	// prometheus server
	flag.StringVar(&PrometheusAddr, "prom-addr", PrometheusAddr, "prometheus listen address")
	flag.IntVar(&PrometheusPort, "prom-port", PrometheusPort, "prometheus listen port")
	flag.BoolFunc("enable-prometheus", "start prometheus", func(s string) error {
		go startPrometheusListener()
		return nil
	})

	flag.BoolFunc("version", "print version", func(s string) error {
		fmt.Println(VERSION)
		os.Exit(0)
		return nil
	})
}
