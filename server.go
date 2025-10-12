package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	Version      string
	Revision     string
	RevisionTime string
)

var (
	// goneypot configuration
	Addr           = "0.0.0.0"
	Port           = 2222
	PrivateKeyFile = "id_rsa"
	LoggingRoot    = ""
	ServerVersion  = "SSH-2.0-OpenSSH_9.9"
	Prompt         = "user@server:~$ "
	Banner         = ""
	CredsFile      = ""
	credentials    = map[string]string{}
	DisableLogin   = false
)

func loadCredentials(credsFile string) error {
	f, err := os.Open(credsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sline := strings.SplitN(scanner.Text(), ":", 2)
		if len(sline) != 2 {
			continue
		}
		username := sline[0]
		password := sline[1]
		if _, found := credentials[username]; found {
			logger.Printf("found duplicate credentials for: %s", username)
			continue
		}
		credentials[username] = password
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func passwordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if DisableLogin {
		return nil, fmt.Errorf("login disabled")
	}

	// if there is no credentials then everyone can log in
	if len(credentials) == 0 {
		return nil, nil
	}

	password, found := credentials[c.User()]
	if !found {
		return nil, fmt.Errorf("user not found %q", c.User())
	}
	if password != string(pass) {
		return nil, fmt.Errorf("wrong password for %q", c.User())
	}
	return nil, nil
}

func getGoneypotListener() net.Listener {
	var listener net.Listener = nil
	var err error
	if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) {
		// systemd file listener for systemd.socket
		if os.Getenv("LISTEN_FDS") != "2" {
			logger.Fatal("LISTEN_FDS should be 2, service expected 2 sockets")
		}
		names := strings.Split(os.Getenv("LISTEN_FDNAMES"), ":")
		for i, name := range names {
			if name == "goneypot" {
				f := os.NewFile(uintptr(i+3), "goneypot port")
				listener, err = net.FileListener(f)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Printf("goneypot listening on systemd socket")
				return listener
			}
		}
		logger.Fatal("no socket listener found for goneypot")
	} else {
		// port bind
		addr := fmt.Sprintf("%s:%d", Addr, Port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Printf("goneypot listening on: %s", addr)
	}
	return listener
}

func startHoneypot() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if !DisableCredsLog {
				logRemoteEvent(c.RemoteAddr().String(), fmt.Sprintf("login attempt: %q:%q", c.User(), string(pass)))
				credsLogger.Printf("%s:%s", c.User(), string(pass))
			} else {
				logRemoteEvent(c.RemoteAddr().String(), "login attempt")
			}

			loginAtempts.Add(1)
			sshPerm, err := passwordCallback(c, pass)
			if err != nil {
				loginFailed.Add(1)
			}
			return sshPerm, err
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
		logger.Fatal("failed to load private key: ", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logger.Fatal("failed to parse private key: ", err)
	}
	config.AddHostKey(private)

	if CredsFile != "" {
		err = loadCredentials(CredsFile)
		if err != nil {
			logger.Fatal("failed to load login credentials: ", err)
		}
		logger.Printf("loaded credentials from: %s", CredsFile)
	}

	listener := getGoneypotListener()

	for {
		nConn, err := listener.Accept()
		if err != nil {
			logger.Printf("failed to accept incoming connection: %s", err)
			totalErrors.Add(1)
			continue
		}

		go func() {
			totalConnections.Add(1)
			openedConnections.Add(1)

			remoteAddr := nConn.RemoteAddr().String()
			logRemoteEvent(remoteAddr, "connection opened")

			now := time.Now()
			handleCon(config, nConn, remoteAddr)
			requestDurations.Observe(time.Since(now).Seconds())

			openedConnections.Sub(1)
		}()
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
	// The incoming Request channel must be serviced.
	go func() {
		ssh.DiscardRequests(reqs)
		logRemoteEvent(remoteAddr, "connection closed")
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		handleChannel(newChannel, remoteAddr)
	}
}

func handleChannel(newChannel ssh.NewChannel, remoteAddr string) {
	// Channels have a type, depending on the application level protocol
	// intended. In the case of a shell, the type is "session" and ServerShell
	// may be used to present a simple terminal interface.
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		logRemoteEvent(remoteAddr, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		logRemoteEvent(remoteAddr, fmt.Sprintf("could not accept channel: %v", err))
		return
	}
	defer func() {
		if channel != nil {
			channel.Close()
		}
	}()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env".
	for req := range requests {
		logRemoteEvent(remoteAddr, fmt.Sprintf("req: %s", req.Type))

		switch req.Type {
		case "shell":
			err := req.Reply(req.Type == "shell", nil)
			if err != nil {
				logger.Printf("failed to reply to shell request: %s", err)
				return
			}

			handleShell(channel, remoteAddr)
			return
		case "env":
			logRemoteEvent(remoteAddr, fmt.Sprintf("env: %q", req.Payload))
		case "pty-req":
		case "window-change":
		case "exec":
			totalCommands.Add(1)

			logFileName := path.Join(LoggingRoot, remoteAddr+".log")
			fo, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				logger.Printf("failed to open log file: %s", err)
				totalErrors.Add(1)
				return
			}
			defer func() {
				fo.Close()
			}()

			_, err = fo.Write(req.Payload)
			if err != nil {
				logger.Printf("failed to write exec payload to log file: %s", err)
				totalErrors.Add(1)
				return
			}
			return
		}
	}
}

func handleShell(channel ssh.Channel, remoteAddr string) {
	term := term.NewTerminal(channel, Prompt)

	logFileName := path.Join(LoggingRoot, remoteAddr+".log")
	fo, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		logger.Printf("failed to open log file: %s", err)
		totalErrors.Add(1)
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

		totalCommands.Add(1)

		_, err = fo.WriteString(line + "\n")
		if err != nil {
			logger.Printf("failed to write command to log file: %s", err)
			totalErrors.Add(1)
			return
		}
	}
}
