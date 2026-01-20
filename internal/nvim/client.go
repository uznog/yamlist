package nvim

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"time"
)

// MinSendInterval is the minimum time between cursor updates (throttling)
const MinSendInterval = 50 * time.Millisecond

// Message represents a JSONL message sent/received from Neovim
type Message struct {
	Op   string `json:"op"`
	Line int    `json:"line,omitempty"`
}

// CursorUpdateMsg is a Bubble Tea message sent when Neovim cursor changes
type CursorUpdateMsg struct {
	Line int
}

// Client handles IPC communication with Neovim via Unix socket
type Client struct {
	conn       net.Conn
	socketPath string
	mu         sync.Mutex
	lastSent   time.Time
	closed     bool
	cursorChan chan int
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
		cursorChan: make(chan int, 10), // Buffered channel to prevent blocking
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
	close(c.cursorChan)
	return c.conn.Close()
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c != nil && !c.closed
}

// CursorUpdates returns a channel that receives cursor line numbers from Neovim
func (c *Client) CursorUpdates() <-chan int {
	if c == nil {
		return nil
	}
	return c.cursorChan
}

// StartReader starts a goroutine to receive messages from Neovim
func (c *Client) StartReader() {
	if c == nil {
		return
	}
	go c.readLoop()
}

// readLoop continuously reads messages from Neovim
func (c *Client) readLoop() {
	if c == nil {
		return
	}

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		if c.closed {
			return
		}

		line := scanner.Text()
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		// Handle cursor position updates from Neovim
		if msg.Op == "cursor" {
			select {
			case c.cursorChan <- msg.Line:
			default:
				// Channel full, drop the message to prevent blocking
			}
		}
	}
}

// Receive reads a single JSONL message from Neovim
// This is a blocking call - use StartReader() for continuous reading
func (c *Client) Receive() (*Message, error) {
	if c == nil || c.closed {
		return nil, nil
	}

	scanner := bufio.NewScanner(c.conn)
	if !scanner.Scan() {
		return nil, scanner.Err()
	}

	var msg Message
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
