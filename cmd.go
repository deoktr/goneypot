package main

import (
	"flag"
	"fmt"
	"os"
)

const VERSION = "1.0.0"

func ExecuteCLI() {
	flag.Parse()

	startHoneypot()
}

func init() {
	flag.StringVar(&Addr, "addr", "0.0.0.0", "honeypot listen address")
	flag.StringVar(&Port, "port", "2222", "honeypot listen port")
	flag.StringVar(&PrivateKeyFile, "key", "id_rsa", "private SSH key file")
	flag.StringVar(&Prompt, "prompt", "[user@db01:~]$ ", "shell prompt")

	flag.BoolFunc("version", "print version", func(s string) error {
		fmt.Println(VERSION)
		os.Exit(0)
		return nil
	})
}
