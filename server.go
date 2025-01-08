package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// TODO: add timeouts, to prevent hanging connections

const VERSION = "1.2.0"

var (
	Prompt         = "[user@server:~]$ "
	Addr           = "0.0.0.0"
	Port           = "2222"
	PrivateKeyFile = "id_rsa"
	ServerVersion  = "SSH-2.0-OpenSSH_9.9"
	Banner         = ""
	User           = ""
	Password       = ""
)

func startHoneypot() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			logRemoteEvent(c.RemoteAddr().String(), fmt.Sprintf("login attempt: %q:%q", c.User(), string(pass)))

			if (User != "" && c.User() != User) || (Password != "" && string(pass) != Password) {
				return nil, fmt.Errorf("password rejected for %q", c.User())
			}

			return nil, nil
		},
	}

	if ServerVersion != "" {
		config.ServerVersion = ServerVersion
	}

	if Banner != "" {
		config.BannerCallback = func(conn ssh.ConnMetadata) string { return Banner + "\n" }
	}

	privateBytes, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		log.Fatal("failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("failed to parse private key: ", err)
	}
	config.AddHostKey(private)

	addr := Addr + ":" + Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	log.Printf("listening on: %s", addr)

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %s", err)
		}

		remoteAddr := nConn.RemoteAddr().String()

		logRemoteEvent(remoteAddr, "connection opened")

		go handleCon(config, nConn, remoteAddr)
	}
}

func handleCon(config *ssh.ServerConfig, nConn net.Conn, remoteAddr string) {
	// Before use, a handshake must be performed on the incoming net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		logRemoteEvent(remoteAddr, fmt.Sprintf("failed to handshake: %s", err))
		return
	}
	logRemoteEvent(remoteAddr, "logged in")

	handleChannels(chans, reqs, remoteAddr)
}

func handleChannels(chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, remoteAddr string) {
	var wg sync.WaitGroup
	defer wg.Wait()

	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
		logRemoteEvent(remoteAddr, "connection closed")
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		handleChannel(newChannel, &wg, remoteAddr)
	}
}

func handleChannel(newChannel ssh.NewChannel, wg *sync.WaitGroup, remoteAddr string) {
	// Channels have a type, depending on the application level protocol
	// intended. In the case of a shell, the type is "session" and ServerShell
	// may be used to present a simple terminal interface.
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		logRemoteEvent(remoteAddr, fmt.Sprintf("could not accept channel: %v", err))
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env".
	wg.Add(1)
	go func(in <-chan *ssh.Request) {
		for req := range in {
			logRemoteEvent(remoteAddr, fmt.Sprintf("req %s: %q", req.Type, req.Payload))

			switch req.Type {
			case "shell":
				req.Reply(req.Type == "shell", nil)
			case "env":
			case "pty-req":
			case "window-change":
			case "exec":
			}
		}
		wg.Done()
	}(requests)

	term := term.NewTerminal(channel, Prompt)

	wg.Add(1)
	go func() {
		defer func() {
			channel.Close()
			wg.Done()
		}()
		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}

			// ignore empty commands
			if line == "" {
				continue
			}

			logRemoteEvent(remoteAddr, fmt.Sprintf("cmd: %q", line))
		}
	}()
}

func logRemoteEvent(remoteAddr string, message string) {
	// TODO: create one log file for each remoteAddr
	log.Printf("%s %s", remoteAddr, message)
}
