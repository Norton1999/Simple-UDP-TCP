Chat Application
A simple, scalable chat application written in Go, supporting TCP-based messaging, UDP-based user discovery, and SQLite database integration for persistent user and message storage. The application includes features like private messaging, message history, user authentication, and graceful shutdown.
Features

TCP Messaging: Real-time chat with broadcast and private messages.
UDP User Discovery: Broadcasts online user list to clients.
Database Integration:
SQLite database (chat.db) for storing users and messages.
User authentication with bcrypt-hashed passwords.
Message history with sender, receiver, content, and timestamp.


Commands:
/pm <username> <message>: Send a private message to a specific user.
/history: Display recent chat history (up to 100 messages).
/users: List online users.


Timeout and Heartbeat:
Configurable timeouts for TCP/UDP connections and client dialing.
Heartbeat mechanism (PING/PONG) to detect inactive clients.


Performance Optimization:
Goroutine pool for efficient message broadcasting.
sync.Pool for UDP buffer allocation.


Graceful Shutdown: Handles SIGINT/SIGTERM signals to clean up resources.
Configurability: Supports environment variables for ports, timeouts, etc.

Project Structure
chat/
├── cmd/
│   ├── client/
│   │   └── main.go         // Client entry point
│   └── server/
│       └── main.go         // Server entry point
├── internal/
│   ├── config/
│   │   └── config.go       // Configuration management
│   ├── tcp/
│   │   └── tcp.go          // TCP server and client logic
│   ├── udp/
│   │   └── udp.go          // UDP broadcast and receive logic
│   ├── history/
│   │   └── history.go      // Message history management
│   ├── auth/
│   │   └── auth.go         // User authentication
│   ├── message/
│   │   └── message.go      // Message type and formatting
│   ├── database/
│   │   └── database.go     // SQLite database operations
│   └── pool/
│       └── pool.go         // Goroutine pool for broadcasting
├── pkg/
│   └── logger/
│       └── logger.go       // Logging utility
└── go.mod                  // Go module definition

Prerequisites

Go: Version 1.21 or higher.
Dependencies:
github.com/mattn/go-sqlite3 v1.14.22 (SQLite driver).
golang.org/x/crypto v0.25.0 (bcrypt for password hashing).



Installation

Clone the repository:git clone <repository-url>
cd chat


Install dependencies:go mod tidy



Configuration
The application uses environment variables for configuration. Default values are provided in internal/config/config.go. You can override them by setting environment variables:
export TCP_PORT=":8888"
export UDP_PORT=":9999"
export BROADCAST_ADDR="255.255.255.255:9999"
export TCP_TIMEOUT="30s"
export UDP_TIMEOUT="5s"
export DIAL_TIMEOUT="10s"
export BROADCAST_INTERVAL="5s"
export HEARTBEAT_INTERVAL="15s"

Usage

Run the Server:
cd cmd/server
go run .


The server starts on TCP port 8888 (chat) and UDP port 9999 (user list broadcast).
A SQLite database (chat.db) is automatically created in the project root to store users and messages.


Run the Client:
cd cmd/client
go run .


Enter a username and password when prompted.
First-time login registers the user (password hashed and stored in chat.db).
Subsequent logins verify credentials against the database.
Type messages to broadcast to all users or use commands:
/pm <username> <message>: Send a private message.
/history: View recent messages (up to 100).
/users: List online users.




Database:

The chat.db file contains two tables:
users: Stores username (TEXT, PRIMARY KEY) and password_hash (TEXT).
messages: Stores id (INTEGER, PRIMARY KEY, AUTOINCREMENT), from_username (TEXT), to_username (TEXT, empty for broadcast), content (TEXT), timestamp (TEXT, UTC format 2006-01-02 15:04:05).


To inspect the database:sqlite3 chat.db
SELECT * FROM users;
SELECT * FROM messages;





Example Interaction

Start the server:cd cmd/server
go run .


Start multiple clients in separate terminals:cd cmd/client
go run .


In each client:
Enter a username (e.g., alice) and password (e.g., pass123).
Send messages:Message: Hello, everyone!
Message: /pm bob Hi, private message!
Message: /history
Message: /




