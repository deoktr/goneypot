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
	flag.StringVar(&ServerVersion, "serverversion", ServerVersion, "ssh server version")
	flag.StringVar(&Banner, "banner", Banner, "ssh banner")
	flag.StringVar(&User, "user", User, "auth username, by default any")
	flag.StringVar(&Password, "pass", Password, "auth password, by default any")

	flag.BoolFunc("version", "print version", func(s string) error {
		fmt.Println(VERSION)
		os.Exit(0)
		return nil
	})
}
