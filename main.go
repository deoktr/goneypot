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

func main() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			log.Printf("login attempt: %s:%s", c.User(), string(pass))
			if c.User() == "user" && string(pass) == "tiger" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile("id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	config.AddHostKey(private)

	addr := "0.0.0.0:2222"
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
		log.Printf("connection opened from: %s", nConn.RemoteAddr())

		go handleCon(config, nConn)
	}
}

func handleCon(config *ssh.ServerConfig, nConn net.Conn) {
	// Before use, a handshake must be performed on the incoming net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Printf("failed to handshake: %s", err)
		return
	}
	log.Printf("logged in")

	handleChannels(chans, reqs)
}

func handleChannels(chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	var wg sync.WaitGroup
	defer wg.Wait()

	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
		log.Println("connection closed")
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		handleChannel(newChannel, &wg)
	}
}

func handleChannel(newChannel ssh.NewChannel, wg *sync.WaitGroup) {
	// Channels have a type, depending on the application level protocol
	// intended. In the case of a shell, the type is "session" and ServerShell
	// may be used to present a simple terminal interface.
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Fatalf("Could not accept channel: %v", err)
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env".
	wg.Add(1)
	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "shell":
				req.Reply(req.Type == "shell", nil)
			case "env":
				// Save environment variables for later use
			case "pty-req":
				// Set the TTY flag on the Docker client to true later
			case "window-change":
				// Use the ContainerResize method on the Docker client later
			case "exec":
				// Create a container and run it
			}
		}
		wg.Done()
	}(requests)

	term := term.NewTerminal(channel, "> ")

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
			fmt.Println(line)
		}
	}()
}
