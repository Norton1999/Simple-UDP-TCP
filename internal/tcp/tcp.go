package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"chat/internal/config"
	"chat/internal/history"
	"chat/internal/message"
	"chat/internal/pool"
	"chat/pkg/logger"
)

// Errors define custom error types
var (
	ErrUsernameTaken = errors.New("ERR001: username already taken")
	ErrAuthFailed    = errors.New("ERR002: authentication failed")
	ErrInvalidCommand = errors.New("ERR003: invalid command")
)

// Server manages TCP connections
type Server struct {
	cfg      config.Config
	logger   *logger.Logger
	history  *history.History
	auth     *auth.AuthManager
	listener net.Listener
	users    map[string]net.Conn
	usersMu  sync.Mutex
	msgChan  chan message.Message
	done     chan struct{}
	pool     *pool.Pool
}

// NewServer creates a new TCP server
func NewServer(cfg config.Config, logger *logger.Logger, hist *history.History, auth *auth.AuthManager, pool *pool.Pool) *Server {
	return &Server{
		cfg:     cfg,
		logger:  logger,
		history: hist,
		auth:    auth,
		users:   make(map[string]net.Conn),
		msgChan: make(chan message.Message, 100),
		done:    make(chan struct{}),
		pool:    pool,
	}
}

// Start runs the TCP server
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.cfg.TCPPort)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	s.logger.Info("TCP server started on %s", s.cfg.TCPPort)

	// Start message broadcasting
	go s.broadcastMessages()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				s.logger.Error("Accept error: %v", err)
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

// Shutdown closes the TCP server
func (s *Server) Shutdown() {
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
	s.usersMu.Lock()
	for _, conn := range s.users {
		conn.Close()
	}
	s.usersMu.Unlock()
}

// handleConnection processes a single TCP connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set timeouts
	conn.SetReadDeadline(time.Now().Add(s.cfg.TCPTimeout))
	conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))

	// Read username and password
	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		s.logger.Error("Failed to read username: %v", err)
		return
	}
	username = strings.TrimSpace(username)
	password, err := reader.ReadString('\n')
	if err != nil {
		s.logger.Error("Failed to read password: %v", err)
		return
	}
	password = strings.TrimSpace(password)

	// Authenticate user
	if !s.auth.Authenticate(username, password) {
		conn.Write([]byte(ErrAuthFailed.Error() + "\n"))
		return
	}

	// Register user
	s.usersMu.Lock()
	if _, exists := s.users[username]; exists {
		conn.Write([]byte(ErrUsernameTaken.Error() + "\n"))
		s.usersMu.Unlock()
		return
	}
	s.users[username] = conn
	s.usersMu.Unlock()

	// Send history messages
	for _, msg := range s.history.GetAll() {
		conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))
		conn.Write([]byte(msg + "\n"))
	}

	// Broadcast user joined
	joinedMsg := message.NewSystemMessage(fmt.Sprintf("%s joined the chat", username))
	s.msgChan <- joinedMsg
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	s.history.Add(joinedMsg.String(), "", "", timestamp)  // System message, from and to empty

	// Start heartbeat
	go s.heartbeat(conn, username)

	// Handle messages
	for {
		conn.SetReadDeadline(time.Now().Add(s.cfg.TCPTimeout))
		input, err := reader.ReadString('\n')
		if err != nil {
			s.usersMu.Lock()
			delete(s.users, username)
			s.usersMu.Unlock()
			leftMsg := message.NewSystemMessage(fmt.Sprintf("%s left the chat", username))
			s.msgChan <- leftMsg
			timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
			s.history.Add(leftMsg.String(), "", "", timestamp)  // System message
			s.logger.Info("User %s disconnected: %v", username, err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			if err := s.processInput(username, input); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			}
		}
	}
}

// processInput handles user input (commands or messages)
func (s *Server) processInput(username, input string) error {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	if strings.HasPrefix(input, "/") {
		return s.handleCommand(username, input, timestamp)
	}
	msg := message.NewUserMessage(username, input)
	s.msgChan <- msg
	s.history.Add(msg.String(), username, "", timestamp)  // Broadcast, to empty
	return nil
}

// handleCommand processes user commands
func (s *Server) handleCommand(username, input, timestamp string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ErrInvalidCommand
	}

	switch parts[0] {
	case "/pm":
		if len(parts) < 3 {
			return fmt.Errorf("ERR004: /pm requires username and message")
		}
		target := parts[1]
		content := strings.Join(parts[2:], " ")
		msg := message.NewPrivateMessage(username, target, content)
		s.msgChan <- msg
		s.history.Add(msg.String(), username, target, timestamp)
	case "/history":
		s.usersMu.Lock()
		if conn, exists := s.users[username]; exists {
			for _, msg := range s.history.GetAll() {
				conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))
				conn.Write([]byte(msg + "\n"))
			}
		}
		s.usersMu.Unlock()
	case "/users":
		s.usersMu.Lock()
		userList := strings.Join(s.GetUsers(), ", ")
		if conn, exists := s.users[username]; exists {
			conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))
			conn.Write([]byte(fmt.Sprintf("Online users: %s\n", userList)))
		}
		s.usersMu.Unlock()
	default:
		return fmt.Errorf("ERR005: unknown command %s", parts[0])
	}
	return nil
}

// broadcastMessages broadcasts messages to users
func (s *Server) broadcastMessages() {
	for {
		select {
		case msg := <-s.msgChan:
			s.pool.Submit(func() {
				s.usersMu.Lock()
				defer s.usersMu.Unlock()
				for username, conn := range s.users {
					if msg.Type == message.TypePrivate && msg.Target != username && msg.From != username {
						continue
					}
					conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))
					_, err := conn.Write([]byte(msg.String() + "\n"))
					if err != nil {
						s.logger.Error("Failed to send to %s: %v", username, err)
					}
				}
			})
		case <-s.done:
			return
		}
	}
}

// heartbeat sends periodic pings to detect inactive clients
func (s *Server) heartbeat(conn net.Conn, username string) {
	ticker := time.NewTicker(s.cfg.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(s.cfg.TCPTimeout))
			if _, err := conn.Write([]byte("PING\n")); err != nil {
				s.usersMu.Lock()
				delete(s.users, username)
				s.usersMu.Unlock()
				leftMsg := message.NewSystemMessage(fmt.Sprintf("%s left the chat (timeout)", username))
				s.msgChan <- leftMsg
				timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
				s.history.Add(leftMsg.String(), "", "", timestamp)
				s.logger.Info("User %s timed out", username)
				return
			}
		case <-s.done:
			return
		}
	}
}

// GetUsers returns the list of online users
func (s *Server) GetUsers() []string {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()
	var userList []string
	for username := range s.users {
		userList = append(userList, username)
	}
	return userList
}

// Client manages TCP client connection
type Client struct {
	cfg      config.Config
	logger   *logger.Logger
	conn     net.Conn
	username string
}

// NewClient creates a new TCP client
func NewClient(cfg config.Config, logger *logger.Logger, username, password string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", cfg.TCPAddr(), cfg.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	// Send username and password
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send username: %v", err)
	}
	_, err = conn.Write([]byte(password + "\n"))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send password: %v", err)
	}

	// Check authentication response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read auth response: %v", err)
	}
	if strings.Contains(response, "ERR002") {
		conn.Close()
		return nil, ErrAuthFailed
	}

	return &Client{
		cfg:      cfg,
		logger:   logger,
		conn:     conn,
		username: username,
	}, nil
}

// Send sends a message to the server
func (c *Client) Send(msg string) error {
	c.conn.SetWriteDeadline(time.Now().Add(c.cfg.TCPTimeout))
	_, err := c.conn.Write([]byte(msg + "\n"))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

// Receive handles incoming messages
func (c *Client) Receive() error {
	reader := bufio.NewReader(c.conn)
	for {
		c.conn.SetReadDeadline(time.Now().Add(c.cfg.TCPTimeout))
		msg, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("server connection lost: %v", err)
		}
		if strings.TrimSpace(msg) == "PING" {
			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.TCPTimeout))
			c.conn.Write([]byte("PONG\n"))
			continue
		}
		fmt.Printf("\n%s", msg)
		fmt.Print("Message: ")
	}
}

// Close closes the client connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}