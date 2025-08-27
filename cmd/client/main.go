package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"chat/internal/config"
	"chat/internal/tcp"
	"chat/internal/udp"
	"chat/pkg/logger"
)

// main starts the client
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New("client")

	// Get username and password
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Start TCP client
	tcpClient, err := tcp.NewClient(cfg, log, username, password)
	if err != nil {
		log.Fatal("Failed to start TCP client: %v", err)
	}
	defer tcpClient.Close()

	// Start UDP receiver
	udpReceiver := udp.NewReceiver(cfg, log)
	go func() {
		if err := udpReceiver.Start(); err != nil {
			log.Error("UDP receiver failed: %v", err)
		}
	}()

	// Start receiving messages
	go func() {
		if err := tcpClient.Receive(); err != nil {
			log.Error("Failed to receive messages: %v", err)
			os.Exit(1)
		}
	}()

	// Handle user input
	for {
		fmt.Print("Message: ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Failed to read input: %v", err)
			return
		}
		msg = strings.TrimSpace(msg)
		if msg != "" {
			if err := tcpClient.Send(msg); err != nil {
				log.Error("Failed to send message: %v", err)
				return
			}
		}
	}
}