package udp

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"chat/internal/config"
	"chat/pkg/logger"
)

// Broadcaster manages UDP broadcasts
type Broadcaster struct {
	cfg       config.Config
	logger    *logger.Logger
	conn      *net.UDPConn
	getUsers  func() []string
	done      chan struct{}
	pool      *sync.Pool
}

// NewBroadcaster creates a new UDP broadcaster
func NewBroadcaster(cfg config.Config, logger *logger.Logger) *Broadcaster {
	return &Broadcaster{
		cfg:    cfg,
		logger: logger,
		done:   make(chan struct{}),
		pool:   &sync.Pool{New: func() interface{} { return make([]byte, 1024) }},
	}
}

// SetGetUsers sets the function to get user list
func (b *Broadcaster) SetGetUsers(f func() []string) {
	b.getUsers = f
}

// Start runs the UDP broadcaster
func (b *Broadcaster) Start() error {
	addr, err := net.ResolveUDPAddr("udp", b.cfg.UDPAddr())
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %v", err)
	}
	b.conn = conn

	b.logger.Info("UDP broadcaster started on %s", b.cfg.UDPAddr())

	for {
		select {
		case <-b.done:
			return nil
		default:
			if b.getUsers == nil {
				continue
			}
			userList := strings.Join(b.getUsers(), ",")
			buf := b.pool.Get().([]byte)
			defer b.pool.Put(buf)
			copy(buf, []byte(userList))
			b.conn.SetWriteDeadline(time.Now().Add(b.cfg.UDPTimeout))
			_, err := b.conn.Write(buf[:len(userList)])
			if err != nil {
				b.logger.Error("UDP broadcast failed: %v", err)
			}
			time.Sleep(b.cfg.BroadcastInterval)
		}
	}
}

// Shutdown closes the UDP broadcaster
func (b *Broadcaster) Shutdown() {
	close(b.done)
	if b.conn != nil {
		b.conn.Close()
	}
}

// Receiver manages UDP user list reception
type Receiver struct {
	cfg    config.Config
	logger *logger.Logger
	conn   *net.UDPConn
	done   chan struct{}
	pool   *sync.Pool
}

// NewReceiver creates a new UDP receiver
func NewReceiver(cfg config.Config, logger *logger.Logger) *Receiver {
	return &Receiver{
		cfg:    cfg,
		logger: logger,
		done:   make(chan struct{}),
		pool:   &sync.Pool{New: func() interface{} { return make([]byte, 1024) }},
	}
}

// Start runs the UDP receiver
func (r *Receiver) Start() error {
	addr, err := net.ResolveUDPAddr("udp", r.cfg.UDPPort)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen UDP: %v", err)
	}
	r.conn = conn

	r.logger.Info("UDP receiver started on %s", r.cfg.UDPPort)

	for {
		select {
		case <-r.done:
			return nil
		default:
			buf := r.pool.Get().([]byte)
			r.conn.SetReadDeadline(time.Now().Add(r.cfg.UDPTimeout))
			n, _, err := r.conn.ReadFromUDP(buf)
			if err != nil {
				r.logger.Error("UDP read failed: %v", err)
				r.pool.Put(buf)
				continue
			}
			userList := string(buf[:n])
			fmt.Printf("\nOnline users: %s\n", userList)
			fmt.Print("Message: ")
			r.pool.Put(buf)
		}
	}
}

// Shutdown closes the UDP receiver
func (r *Receiver) Shutdown() {
	close(r.done)
	if r.conn != nil {
		r.conn.Close()
	}
}