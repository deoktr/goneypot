package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// TODO: add timeouts, to prevent hanging connections

const VERSION = "1.4.0"

var (
	Addr           = "0.0.0.0"
	Port           = "2222"
	PrivateKeyFile = "id_rsa"
	LoggingRoot    = ""
	ServerVersion  = "SSH-2.0-OpenSSH_9.9"
	Prompt         = "[user@server:~]$ "
	Banner         = ""
	User           = ""
	Password       = ""

	// remote event logger to stdout
	remoteLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
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
			switch req.Type {
			case "shell":
				req.Reply(req.Type == "shell", nil)
			case "env":
				logRemoteEvent(remoteAddr, fmt.Sprintf("env: %q", req.Payload))
			case "pty-req":
			case "window-change":
			case "exec":
				// log exec to files like regular commands
				logFileName := path.Join(LoggingRoot, remoteAddr+".log")
				fo, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
				if err != nil {
					log.Printf("failed to open log file: %s", err)
					return
				}
				defer func() {
					fo.Close()
				}()

				_, err = fo.Write(req.Payload)
				if err != nil {
					log.Printf("failed to write exec payload to log file: %s", err)
					return
				}
				fo.Close()
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

		logFileName := path.Join(LoggingRoot, remoteAddr+".log")
		fo, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("failed to open log file: %s", err)
			return
		}
		defer func() {
			fo.Close()
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

			_, err = fo.WriteString(line + "\n")
			if err != nil {
				log.Printf("failed to write command to log file: %s", err)
				return
			}
		}
	}()
}

func logRemoteEvent(remoteAddr string, message string) {
	remoteLogger.Printf("%s %s", remoteAddr, message)
}
