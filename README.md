# Chat Application

A simple, scalable chat application written in Go, supporting TCP-based messaging, UDP-based user discovery, and SQLite database integration for persistent user and message storage. The application includes features like private messaging, message history, user authentication, and graceful shutdown.

## Features

- **TCP Messaging**: Real-time chat with broadcast and private messages.
- **UDP User Discovery**: Broadcasts online user list to clients.
- **Database Integration**:
  - SQLite database (`chat.db`) for storing users and messages.
  - User authentication with bcrypt-hashed passwords.
  - Message history with sender, receiver, content, and timestamp.
- **Commands**:
  - `/pm <username> <message>`: Send a private message to a specific user.
  - `/history`: Display recent chat history (up to 100 messages).
  - `/users`: List online users.
- **Timeout and Heartbeat**:
  - Configurable timeouts for TCP/UDP connections and client dialing.
  - Heartbeat mechanism (PING/PONG) to detect inactive clients.
- **Performance Optimization**:
  - Goroutine pool for efficient message broadcasting.
  - `sync.Pool` for UDP buffer allocation.
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals to clean up resources.
- **Configurability**: Supports environment variables for ports, timeouts, etc.

## Project Structure

```plaintext
chat/
├── cmd/
│   ├── client/
│   │   └── main.go         // Client entry point
│   └── server/
│       └── main.go         // Server entry point
├── internal/
│   ├── auth/
│   │   └── auth.go         // User authentication
│   ├── config/
│   │   └── config.go       // Configuration management
│   ├── database/
│   │   └── database.go     // SQLite database operations
│   ├── history/
│   │   └── history.go      // Message history management
│   ├── message/
│   │   └── message.go      // Message type and formatting
│   ├── pool/
│   │   └── pool.go         // Goroutine pool for broadcasting
│   ├── tcp/
│   │   └── tcp.go          // TCP server and client logic
│   └── udp/
│       └── udp.go          // UDP broadcast and receive logic
├── pkg/
│   └── logger/
│       └── logger.go       // Logging utility
├── chat.db                 // SQLite database file (created on server start)
├── go.mod                  // Go module definition
└── README.md               // Project documentation
```

## Prerequisites

- **Go**: Version 1.21 or higher.
- **Dependencies**:
  - `github.com/mattn/go-sqlite3 v1.14.22` (SQLite driver).
  - `golang.org/x/crypto v0.25.0` (bcrypt for password hashing).

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd chat
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Configuration

The application uses environment variables for configuration. Default values are defined in `internal/config/config.go`. Override them by setting environment variables:

```bash
export TCP_PORT=":8888"
export UDP_PORT=":9999"
export BROADCAST_ADDR="255.255.255.255:9999"
export TCP_TIMEOUT="30s"
export UDP_TIMEOUT="5s"
export DIAL_TIMEOUT="10s"
export BROADCAST_INTERVAL="5s"
export HEARTBEAT_INTERVAL="15s"
```

## Usage

1. **Run the Server**:
   ```bash
   cd cmd/server
   go run .
   ```
   - The server starts on TCP port 8888 (chat) and UDP port 9999 (user list broadcast).
   - A SQLite database (`chat.db`) is automatically created in the project root to store users and messages.

2. **Run the Client**:
   ```bash
   cd cmd/client
   go run .
   ```
   - Enter a username and password when prompted.
   - First-time login registers the user (password hashed and stored in `chat.db`).
   - Subsequent logins verify credentials against the database.
   - Send messages or use commands:
     ```plaintext
     Message: Hello, everyone!
     Message: /pm bob Hi, private message!
     Message: /history
     Message: /users
     ```

3. **Database Inspection**:
   - The `chat.db` file contains two tables:
     - `users`: Stores `username` (TEXT, PRIMARY KEY) and `password_hash` (TEXT).
     - `messages`: Stores `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT), `from_username` (TEXT), `to_username` (TEXT, empty for broadcast), `content` (TEXT), `timestamp` (TEXT, UTC format `2006-01-02 15:04:05`).
   - Inspect the database using SQLite:
     ```bash
     sqlite3 chat.db
     .tables
     SELECT * FROM users;
     SELECT * FROM messages;
     ```

## Example Interaction

1. Start the server:
   ```bash
   cd cmd/server
   go run .
   ```

2. Start multiple clients in separate terminals:
   ```bash
   cd cmd/client
   go run .
   ```

3. In each client:
   - Enter a username (e.g., `alice`) and password (e.g., `pass123`).
   - Send messages or commands:
     ```plaintext
     Message: Hello, everyone!
     Message: /pm bob Hi, private message!
     Message: /history
     Message: /users
     ```

4. Messages are saved to `chat.db` with sender, receiver, content, and timestamp.

## Notes

- **Ports**: Ensure ports 8888 (TCP) and 9999 (UDP) are free.
- **Database**: The `chat.db` file persists data across server restarts. Delete it to reset.
- **Security**: Passwords are hashed with bcrypt (default cost). For production, consider increasing bcrypt cost or adding TLS.
- **Scalability**: In-memory history is capped at 100 messages, but the database stores all messages. Add a cleanup mechanism for old messages if needed.
- **File Encoding**: Ensure files use UTF-8 encoding and Unix-style line endings (LF) for GitHub compatibility.

## Extending the Application

- **Full History Query**: Add a `/fullhistory` command to query all messages from the database.
- **Encryption**: Implement TLS for TCP connections to secure messages.
- **Message Cleanup**: Add a function to delete old messages from the database.
- **GUI**: Create a web-based or desktop client using a framework like `fyne` or `React`.

## Troubleshooting

- **Markdown Rendering Issues**: Ensure code blocks use proper language tags (e.g., ```bash, ```go, ```plaintext).
- **Database Errors**: Check if `chat.db` is writable and SQLite is installed.
- **Connection Issues**: Verify ports are open and firewall settings allow TCP/UDP traffic.

## License

MIT License. See [LICENSE](LICENSE) for details.
