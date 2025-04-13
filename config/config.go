package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	WebSocketPort int
	TCPPort       int
}

// NewConfig creates a new configuration instance with default values
func NewConfig() *Config {
	return &Config{
		WebSocketPort: 8080,
		TCPPort:       8081,
	}
}

// LoadConfig loads configuration from flags and environment variables
func LoadConfig() *Config {
	config := NewConfig()

	// Define flags
	wsPort := flag.Int("ws-port", 8080, "WebSocket server port")
	tcpPort := flag.Int("tcp-port", 8081, "TCP server port")

	// Parse flags
	flag.Parse()

	// Get environment variables
	wsPortEnv := os.Getenv("PYSTRAND_WS_PORT")
	tcpPortEnv := os.Getenv("PYSTRAND_TCP_PORT")

	// Use environment variables if set, otherwise use flag values
	if wsPortEnv != "" {
		if port, err := strconv.Atoi(wsPortEnv); err == nil {
			config.WebSocketPort = port
		}
	} else {
		config.WebSocketPort = *wsPort
	}

	if tcpPortEnv != "" {
		if port, err := strconv.Atoi(tcpPortEnv); err == nil {
			config.TCPPort = port
		}
	} else {
		config.TCPPort = *tcpPort
	}

	return config
}
