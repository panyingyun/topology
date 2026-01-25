// Package sshtunnel provides SSH local port forwarding so database connections
// can reach a host only accessible via a jump server.
package sshtunnel

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds SSH jump server and optional auth (password or private key).
type Config struct {
	SSHHost     string
	SSHPort     int
	SSHUser     string
	SSHPassword string
	SSHKey      string // PEM-encoded private key; optional passphrase in SSHPassword when key is encrypted
	DBHost      string
	DBPort      int
}

type tunnel struct {
	listener net.Listener
	client   *ssh.Client
	port     int
	done     chan struct{}
}

var (
	mu      sync.RWMutex
	tunnels = make(map[string]*tunnel)
)

func sshPort(port int) int {
	if port <= 0 {
		return 22
	}
	return port
}

// GetOrStart starts an SSH tunnel for connID if not already running, and returns the local port to connect to.
// DB traffic should connect to 127.0.0.1:localPort to reach DBHost:DBPort via the jump server.
func GetOrStart(connID string, cfg Config) (localPort int, err error) {
	mu.Lock()
	defer mu.Unlock()
	if t, ok := tunnels[connID]; ok {
		// Verify tunnel is still alive
		if _, _, err := t.client.SendRequest("keepalive@openssh.com", true, nil); err == nil {
			return t.port, nil
		}
		_ = t.listener.Close()
		_ = t.client.Close()
		delete(tunnels, connID)
	}

	auth, err := buildAuth(cfg.SSHPassword, cfg.SSHKey)
	if err != nil {
		return 0, fmt.Errorf("ssh auth: %w", err)
	}

	sshAddr := net.JoinHostPort(cfg.SSHHost, strconv.Itoa(sshPort(cfg.SSHPort)))
	clientConfig := &ssh.ClientConfig{
		User:            cfg.SSHUser,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // optional: use ssh.FixedHostKey for production
		Timeout:         15 * time.Second,
	}
	client, err := ssh.Dial("tcp", sshAddr, clientConfig)
	if err != nil {
		return 0, fmt.Errorf("ssh dial: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = client.Close()
		return 0, fmt.Errorf("listen: %w", err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port
	dbAddr := net.JoinHostPort(cfg.DBHost, strconv.Itoa(cfg.DBPort))
	done := make(chan struct{})

	go acceptAndForward(listener, client, dbAddr, done)

	tunnels[connID] = &tunnel{listener: listener, client: client, port: port, done: done}
	return port, nil
}

func buildAuth(password, privateKeyPEM string) ([]ssh.AuthMethod, error) {
	if privateKeyPEM != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
		if err != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKeyPEM), []byte(password))
			if err != nil {
				return nil, fmt.Errorf("parse private key: %w", err)
			}
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	}
	if password != "" {
		return []ssh.AuthMethod{ssh.Password(password)}, nil
	}
	return nil, fmt.Errorf("no ssh auth: set password or privateKey")
}

func acceptAndForward(listener net.Listener, client *ssh.Client, remoteAddr string, _ chan struct{}) {
	for {
		localConn, err := listener.Accept()
		if err != nil {
			return
		}
		remoteConn, err := client.Dial("tcp", remoteAddr)
		if err != nil {
			_ = localConn.Close()
			continue
		}
		go copyBoth(localConn, remoteConn)
	}
}

func copyBoth(a, b net.Conn) {
	defer a.Close()
	defer b.Close()
	done := make(chan struct{}, 1)
	go func() {
		_, _ = io.Copy(b, a)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(a, b)
		done <- struct{}{}
	}()
	<-done
}

// Stop closes the SSH tunnel for the given connID.
func Stop(connID string) {
	mu.Lock()
	defer mu.Unlock()
	t, ok := tunnels[connID]
	if !ok {
		return
	}
	close(t.done)
	_ = t.listener.Close()
	_ = t.client.Close()
	delete(tunnels, connID)
}
