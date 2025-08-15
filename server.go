package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// User management
var (
	users   = make(map[string]net.Conn)
	usersMu sync.Mutex
)

// Server configuration
const (
	tcpPort       = ":8888"
	udpPort       = ":9999"
	broadcastAddr = "255.255.255.255:9999"
)

func main() {
	// Start TCP server
	go startTCPServer()
	// Start UDP broadcast
	go startUDPBroadcast()

	// Keep server running
	select {}
}

// Start TCP server
func startTCPServer() {
	listener, err := net.Listen("tcp", tcpPort)
	if err != nil {
		fmt.Printf("TCP server error: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("TCP server started on %s\n", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		go handleTCPConnection(conn)
	}
}

// Handle TCP connection
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// Read username
	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Read username error: %v\n", err)
		return
	}
	username = strings.TrimSpace(username)

	// Register user
	usersMu.Lock()
	if _, exists := users[username]; exists {
		conn.Write([]byte("Username already taken\n"))
		usersMu.Unlock()
		return
	}
	users[username] = conn
	usersMu.Unlock()

	// Broadcast user joined
	broadcastMessage(fmt.Sprintf("%s joined the chat", username))

	// Handle messages
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			usersMu.Lock()
			delete(users, username)
			usersMu.Unlock()
			broadcastMessage(fmt.Sprintf("%s left the chat", username))
			return
		}
		msg = strings.TrimSpace(msg)
		if msg != "" {
			broadcastMessage(fmt.Sprintf("%s: %s", username, msg))
		}
	}
}

// Broadcast message to all users
func broadcastMessage(msg string) {
	usersMu.Lock()
	defer usersMu.Unlock()

	for _, conn := range users {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
		}
	}
}

// Start UDP broadcast
func startUDPBroadcast() {
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		fmt.Printf("Resolve UDP addr error: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("UDP dial error: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		usersMu.Lock()
		userList := strings.Join(getUserList(), ",")
		usersMu.Unlock()

		_, err := conn.Write([]byte(userList))
		if err != nil {
			fmt.Printf("UDP broadcast error: %v\n", err)
		}
		time.Sleep(5 * time.Second)
	}
}

// Get list of online users
func getUserList() []string {
	var userList []string
	for username := range users {
		userList = append(userList, username)
	}
	return userList
}