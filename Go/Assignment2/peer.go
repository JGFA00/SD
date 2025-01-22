package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Peer struct holds information about the peer's host, port, and neighbors
type Peer struct {
	Host      string
	Port      int
	Neighbors map[string]time.Time // [IP] -> timestamp
	mu        sync.Mutex           // Protects access to Neighbors
}

// NewPeer creates a new Peer with the given host and port
func NewPeer(host string, port int) *Peer {
	return &Peer{
		Host:      host,
		Port:      port,
		Neighbors: make(map[string]time.Time),
	}
}

// StartServer starts the peer's server to listen for incoming connections
func (p *Peer) StartServer() {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server on %s: %v", addr, err)
	}
	defer listener.Close()
	log.Printf("Peer server listening on %s...", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go p.handleConnection(conn)
	}
}

// handleConnection processes an incoming connection and updates the neighbor map
func (p *Peer) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()
		log.Printf("Received data: %s", formatNeighborData(data))
		p.updateNeighbors(data)
	}
}

// updateNeighbors updates the Neighbors map with new entries from received data
func (p *Peer) updateNeighbors(data string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entries := strings.Split(data, ";")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ",")
		if len(parts) == 2 {
			ip := parts[0]
			timestamp, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				p.Neighbors[ip] = time.Unix(timestamp, 0)
			}
		}
	}
}

// disseminateNeighbors sends the Neighbors map to all known neighbors
func (p *Peer) disseminateNeighbors() {
	p.mu.Lock()
	data := ""
	for ip, timestamp := range p.Neighbors {
		data += fmt.Sprintf("%s,%d;", ip, timestamp.Unix())
	}
	p.mu.Unlock()

	for ip := range p.Neighbors {
		go func(ip string) {
			conn, err := net.Dial("tcp", ip)
			if err != nil {
				log.Printf("Failed to connect to neighbor %s: %v", ip, err)
				return
			}
			defer conn.Close()
			fmt.Fprintf(conn, "%s\n", data)
		}(ip)
	}
}

// cleanupNeighbors removes stale neighbors based on a threshold
func (p *Peer) cleanupNeighbors() {
	threshold := 2 * time.Minute // Example: 2-minute threshold
	for {
		time.Sleep(30 * time.Second) // Periodically cleanup
		p.mu.Lock()
		now := time.Now()
		for ip, timestamp := range p.Neighbors {
			if now.Sub(timestamp) > threshold {
				delete(p.Neighbors, ip)
				log.Printf("Removed stale neighbor: %s", ip)
			}
		}
		p.mu.Unlock()
	}
}

// Helper function to convert string to int with error handling
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}
	return i
}

// formatNeighborData converts raw neighbor data into a readable format
func formatNeighborData(data string) string {
	entries := strings.Split(data, ";")
	formatted := []string{}
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ",")
		if len(parts) == 2 {
			ip := parts[0]
			timestamp, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				readableTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
				formatted = append(formatted, fmt.Sprintf("%s (Last updated: %s)", ip, readableTime))
			} else {
				formatted = append(formatted, fmt.Sprintf("%s (Invalid timestamp: %s)", ip, parts[1]))
			}
		}
	}
	return strings.Join(formatted, "; ")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run peer.go <host:port> [<host:port>...]")
	}

	// Parse the first argument as the current peer's address
	currentAddr := os.Args[1]
	parts := strings.Split(currentAddr, ":")
	if len(parts) != 2 {
		log.Fatalf("Invalid address format: %s", currentAddr)
	}

	host := parts[0]
	port := atoi(parts[1])

	peer := NewPeer(host, port)

	// Parse additional arguments as neighbor addresses
	for _, addr := range os.Args[2:] {
		peer.mu.Lock()
		peer.Neighbors[addr] = time.Now()
		peer.mu.Unlock()
	}

	go peer.StartServer()
	go peer.cleanupNeighbors()

	pp := NewPoissonProcess(0.0333, time.Now().UnixNano()) // Poisson process 2 times per minute

	for {
		time.Sleep(time.Duration(pp.TimeForNextEvent()) * time.Second) // Wait based on Poisson interval
		peer.disseminateNeighbors()
		log.Printf("Disseminated neighbors. Current map size: %d", len(peer.Neighbors))
	}
}
