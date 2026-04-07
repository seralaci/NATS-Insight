package nats

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/seralaci/nats-insight/internal/store"
)

type ConnectionStatus struct {
	Connected    bool   `json:"connected"`
	ConnectionID string `json:"connectionId,omitempty"`
	ServerName   string `json:"serverName,omitempty"`
	ServerID     string `json:"serverId,omitempty"`
	ClusterName  string `json:"clusterName,omitempty"`
	Version      string `json:"version,omitempty"`
	RTT          string `json:"rtt,omitempty"`
	Error        string `json:"error,omitempty"`
}

type Manager struct {
	mu         sync.RWMutex
	conn       *natsclient.Conn
	js         jetstream.JetStream
	connID     string
	monitorURL string
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Connect(c *store.Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Disconnect existing connection if any
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
		m.connID = ""
		m.monitorURL = ""
	}

	opts := []natsclient.Option{
		natsclient.Name("NATS-Insight"),
		natsclient.ReconnectWait(2 * time.Second),
		natsclient.MaxReconnects(-1),
		natsclient.DisconnectErrHandler(func(_ *natsclient.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		natsclient.ReconnectHandler(func(_ *natsclient.Conn) {
			slog.Info("NATS reconnected")
		}),
	}

	switch c.AuthMethod {
	case "username_password":
		opts = append(opts, natsclient.UserInfo(c.Username, c.Password))
	case "token":
		opts = append(opts, natsclient.Token(c.Token))
	case "nkey":
		opt, err := natsclient.NkeyOptionFromSeed(c.NKey)
		if err != nil {
			return fmt.Errorf("invalid nkey: %w", err)
		}
		opts = append(opts, opt)
	case "credentials":
		opts = append(opts, natsclient.UserCredentials(c.CredsFile))
	}

	nc, err := natsclient.Connect(c.URL, opts...)
	if err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}

	m.conn = nc
	m.connID = c.ID
	m.monitorURL = c.MonitorURL

	js, err := jetstream.New(nc)
	if err != nil {
		slog.Warn("JetStream not available", "error", err)
		m.js = nil
	} else {
		m.js = js
	}

	slog.Info("connected to NATS", "url", c.URL, "server", nc.ConnectedServerName())
	return nil
}

func (m *Manager) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
		m.js = nil
		m.connID = ""
		m.monitorURL = ""
		slog.Info("disconnected from NATS")
	}
}

func (m *Manager) Status() ConnectionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.conn == nil || !m.conn.IsConnected() {
		return ConnectionStatus{Connected: false, ConnectionID: m.connID}
	}

	rtt, err := m.conn.RTT()
	rttStr := ""
	if err == nil {
		rttStr = rtt.String()
	}

	return ConnectionStatus{
		Connected:    true,
		ConnectionID: m.connID,
		ServerName:   m.conn.ConnectedServerName(),
		ServerID:     m.conn.ConnectedServerId(),
		ClusterName:  m.conn.ConnectedClusterName(),
		Version:      m.conn.ConnectedServerVersion(),
		RTT:          rttStr,
	}
}

func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn != nil && m.conn.IsConnected()
}

func (m *Manager) ActiveConnectionID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connID
}

// MonitorURL returns the monitoring URL for the active connection.
func (m *Manager) MonitorURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.monitorURL
}

// JetStream returns the JetStream context or an error if not available.
func (m *Manager) JetStream() (jetstream.JetStream, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.js == nil {
		return nil, fmt.Errorf("JetStream not available")
	}
	return m.js, nil
}

// Conn returns the raw NATS connection or an error if not connected.
func (m *Manager) Conn() (*natsclient.Conn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.conn == nil || !m.conn.IsConnected() {
		return nil, fmt.Errorf("not connected to NATS")
	}
	return m.conn, nil
}

// FetchMonitorEndpoint fetches data from a NATS monitoring HTTP endpoint.
func (m *Manager) FetchMonitorEndpoint(endpoint string) (json.RawMessage, error) {
	m.mu.RLock()
	monitorURL := m.monitorURL
	m.mu.RUnlock()

	if monitorURL == "" {
		return nil, fmt.Errorf("no monitor URL configured")
	}

	url := monitorURL + endpoint
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d", endpoint, resp.StatusCode)
	}

	return json.RawMessage(body), nil
}
