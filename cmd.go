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
	flag.StringVar(&Addr, "addr", Addr, "honeypot listen address")
	flag.StringVar(&Port, "port", Port, "honeypot listen port")
	flag.StringVar(&PrivateKeyFile, "key", PrivateKeyFile, "private SSH key file")
	flag.StringVar(&LoggingRoot, "logroot", LoggingRoot, "logging root directory (default current)")
	flag.StringVar(&Prompt, "prompt", Prompt, "shell prompt")
	flag.StringVar(&ServerVersion, "server-version", ServerVersion, "ssh server version")
	flag.StringVar(&Banner, "banner", Banner, "ssh banner")
	flag.StringVar(&User, "user", User, "auth username, by default any")
	flag.StringVar(&Password, "pass", Password, "auth password, by default any")
	flag.BoolVar(&DisableLogin, "disable-login", DisableLogin, "disable login")

	flag.StringVar(&PrometheusAddr, "prom-addr", PrometheusAddr, "prometheus listen address")
	flag.StringVar(&PrometheusPort, "prom-port", PrometheusPort, "prometheus listen port")
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
