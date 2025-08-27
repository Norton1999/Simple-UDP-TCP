package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds server and client configuration
type Config struct {
	TCPPort           string
	UDPPort           string
	BroadcastAddr     string
	TCPTimeout        time.Duration
	UDPTimeout        time.Duration
	DialTimeout       time.Duration
	BroadcastInterval time.Duration
	HeartbeatInterval time.Duration
}

// Load loads configuration from environment variables or defaults
func Load() (Config, error) {
	cfg := Config{
		TCPPort:           getEnv("TCP_PORT", ":8888"),
		UDPPort:           getEnv("UDP_PORT", ":9999"),
		BroadcastAddr:     getEnv("BROADCAST_ADDR", "255.255.255.255:9999"),
		TCPTimeout:        parseDuration(getEnv("TCP_TIMEOUT", "30s")),
		UDPTimeout:        parseDuration(getEnv("UDP_TIMEOUT", "5s")),
		DialTimeout:       parseDuration(getEnv("DIAL_TIMEOUT", "10s")),
		BroadcastInterval: parseDuration(getEnv("BROADCAST_INTERVAL", "5s")),
		HeartbeatInterval: parseDuration(getEnv("HEARTBEAT_INTERVAL", "15s")),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// TCPAddr returns the TCP server address
func (c Config) TCPAddr() string {
	return "localhost" + c.TCPPort
}

// UDPAddr returns the UDP broadcast address
func (c Config) UDPAddr() string {
	return c.BroadcastAddr
}

// Validate checks configuration validity
func (c Config) Validate() error {
	if c.TCPPort == "" || c.UDPPort == "" || c.BroadcastAddr == "" {
		return fmt.Errorf("ports or broadcast address cannot be empty")
	}
	if c.TCPTimeout <= 0 || c.UDPTimeout <= 0 || c.DialTimeout <= 0 || c.BroadcastInterval <= 0 || c.HeartbeatInterval <= 0 {
		return fmt.Errorf("timeouts must be positive")
	}
	return nil
}

// getEnv retrieves environment variable or fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// parseDuration parses duration string or returns default
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}