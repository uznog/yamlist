package nvim

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

// MinSendInterval is the minimum time between cursor updates (throttling)
const MinSendInterval = 50 * time.Millisecond

// Message represents a JSONL message sent to Neovim
type Message struct {
	Op   string `json:"op"`
	Line int    `json:"line,omitempty"`
}

// Client handles IPC communication with Neovim via Unix socket
type Client struct {
	conn       net.Conn
	socketPath string
	mu         sync.Mutex
	lastSent   time.Time
	closed     bool
}

// NewClient creates a new Neovim IPC client
// Returns nil if connection fails (allows standalone mode)
func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       conn,
		socketPath: socketPath,
		lastSent:   time.Time{},
	}, nil
}

// SendCursor sends a cursor position update to Neovim
// Throttled to prevent overwhelming the socket with rapid updates
func (c *Client) SendCursor(line int) error {
	if c == nil || c.closed {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Throttle: skip if last send was too recent
	now := time.Now()
	if now.Sub(c.lastSent) < MinSendInterval {
		return nil
	}

	msg := Message{
		Op:   "cursor",
		Line: line,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send as JSONL (JSON + newline)
	data = append(data, '\n')

	_, err = c.conn.Write(data)
	if err != nil {
		return err
	}

	c.lastSent = now
	return nil
}

// Close closes the connection to Neovim
func (c *Client) Close() error {
	if c == nil || c.closed {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	return c.conn.Close()
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c != nil && !c.closed
}
