# Pystrand-Go

[![Go Report Card](https://goreportcard.com/badge/github.com/pystrand/pystrand-go)](https://goreportcard.com/report/github.com/pystrand/pystrand-go)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A high-performance WebSocket server library written in Go, designed for building real-time applications. Pystrand-Go provides a robust foundation for creating scalable WebSocket-based services with built-in support for room-based communication and message routing.

## 🚀 Features

- **High Performance**: Built with Go for exceptional performance
- **Scalable Architecture**: Stateless design enables horizontal scaling
- **Real-time Communication**: Full-duplex WebSocket communication
- **Room-based Messaging**: Built-in support for room-based communication
- **Security Ready**: Planned support for secure communication (coming soon)
- **Simple Integration**: Easy to set up and use in Go applications
- **TCP Backend Support**: Built-in TCP server for backend communication

## 📦 Installation

```bash
go get github.com/pystrand/pystrand-go
```

## 🛠️ Quick Start

```go
package main

import (
    "github.com/pystrand/pystrand-go/bridge"
)

func main() {
    // Create a new bridge instance
    b := bridge.NewBridge()
    
    // Start the servers
    b.Start()
    
    // Your server is now running!
    // WebSocket server on :8080
    // TCP server on :8081
}
```

## 🔧 Configuration

The server can be configured through command-line flags or environment variables:

### Command-line Flags
```bash
# Start server with custom ports
./your-binary --ws-port=9000 --tcp-port=9001
```

### Environment Variables
```bash
# Set ports using environment variables
export PYSTRAND_WS_PORT=9000
export PYSTRAND_TCP_PORT=9001
./your-binary
```

Default values:
- WebSocket port: 8080
- TCP port: 8081

Note: Environment variables take precedence over command-line flags.

## 📚 Documentation

### Architecture

Pystrand-Go consists of three main components:

1. **WebSocket Server**: Handles client connections and message routing
2. **TCP Server**: Manages communication with backend services
3. **Bridge**: Coordinates between WebSocket and TCP servers

### Message Flow

```
WebSocket Client <-> WebSocket Server <-> Bridge <-> TCP Server <-> Backend Service
```

### Room Management

- Rooms are created dynamically as clients join
- Messages can be broadcast to specific rooms
- Direct messaging to specific clients is supported

## 🔐 Security (Coming Soon)

- Request signing with secret key
- Timestamp-based request validation
- Header-based authentication
- Support for various auth methods (JWT, OAuth, etc.)

## 📈 Scalability

- Stateless design for horizontal scaling
- Vertical scaling support for Go server
- Multiple backend service instances support
- Efficient connection pooling

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Fork the repository
2. Clone your fork
3. Create a new branch
4. Make your changes
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

For support, please open an issue in the GitHub repository.

## 📊 Roadmap

- [ ] Implement security features
- [ ] Add monitoring and metrics
- [ ] Improve error handling
- [ ] Add documentation
- [ ] Add tests
- [ ] Add examples
- [ ] Add benchmarks

## 🙏 Acknowledgments

- [Gorilla WebSocket](https://github.com/gorilla/websocket) for the WebSocket implementation
- The Go community for excellent tools and libraries

## 📄 Version

Current version: 0.0.1 (Beta)
