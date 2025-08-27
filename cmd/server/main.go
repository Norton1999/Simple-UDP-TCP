package main

import (
	"os"
	"os/signal"
	"syscall"

	"chat/internal/auth"
	"chat/internal/config"
	"chat/internal/database"
	"chat/internal/history"
	"chat/internal/pool"
	"chat/internal/tcp"
	"chat/internal/udp"
	"chat/pkg/logger"
)

// main starts the server
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New("server")

	// Initialize database
	db, err := database.New("chat.db")
	if err != nil {
		log.Fatal("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize message history with database
	hist := history.New(100, db)

	// Initialize authentication manager with database
	authMgr := auth.New(db)

	// Initialize goroutine pool
	gPool := pool.New(10)

	// Start TCP server
	tcpServer := tcp.NewServer(cfg, log, hist, authMgr, gPool)
	go func() {
		if err := tcpServer.Start(); err != nil {
			log.Fatal("TCP server failed: %v", err)
		}
	}()

	// Start UDP broadcaster
	udpBroadcaster := udp.NewBroadcaster(cfg, log)
	udpBroadcaster.SetGetUsers(tcpServer.GetUsers)
	go func() {
		if err := udpBroadcaster.Start(); err != nil {
			log.Fatal("UDP broadcaster failed: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server...")
	tcpServer.Shutdown()
	udpBroadcaster.Shutdown()
	gPool.Shutdown()
}