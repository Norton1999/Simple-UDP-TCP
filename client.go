package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Client configuration
const (
	serverAddr = "localhost:8888"
	udpPort    = ":9999"
)

func main() {
	// Get username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading username: %v\n", err)
		return
	}
	username = strings.TrimSpace(username)

	// Connect to TCP server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("TCP connection error: %v\n", err)
		return
	}
	defer conn.Close()

	// Send username to server
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		fmt.Printf("Error sending username: %v\n", err)
		return
	}

	// Start UDP receiver
	go startUDPReceive()

	// Handle server messages
	go handleServerMessages(conn)

	// Handle user input
	for {
		fmt.Print("Message: ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}
		msg = strings.TrimSpace(msg)
		if msg != "" {
			_, err = conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
				return
			}
		}
	}
}

// Handle messages from server
func handleServerMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Server connection lost: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n%s", msg)
		fmt.Print("Message: ")
	}
}

// Start UDP receiver for user list
func startUDPReceive() {
	addr, err := net.ResolveUDPAddr("udp", udpPort)
	if err != nil {
		fmt.Printf("Resolve UDP addr error: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("UDP listen error: %v\n", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("UDP read error: %v\n", err)
			continue
		}
		userList := string(buffer[:n])
		fmt.Printf("\nOnline users: %s\n", userList)
		fmt.Print("Message: ")
	}
}