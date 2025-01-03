package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
    "math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Message represents a network message
type Message struct {
	Content   string
	Timestamp int
	Sender    string
}

// Peer represents a network peer
type Peer struct {
	Host       string
	Port       int
	Neighbors  []string
	mu         sync.Mutex
	Clock      int
	MessageQ   []Message
	Processed  map[string]bool
	ReadyPeers map[string]bool
}

// NewPeer creates a new Peer
func NewPeer(host string, port int, neighbors []string) *Peer {
	return &Peer{
		Host:       host,
		Port:       port,
		Neighbors:  neighbors,
		MessageQ:   make([]Message, 0),
		Processed:  make(map[string]bool),
		ReadyPeers: make(map[string]bool),
	}
}

// Utility to generate a random word
func randomWord() string {
	words := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "kiwi", "lemon"}
	return words[rand.Intn(len(words))]
}

// Utility function to find max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Utility to clean and validate neighbor addresses
func cleanAddress(addr string) string {
	return strings.TrimSpace(addr)
}

func parseNeighbors(neighbors []string) []string {
	validNeighbors := []string{}
	for _, neighbor := range neighbors {
		cleaned := cleanAddress(neighbor)
		if !strings.Contains(cleaned, ":") {
			log.Printf("Skipping invalid neighbor: %s", cleaned)
			continue
		}
		validNeighbors = append(validNeighbors, cleaned)
	}
	return validNeighbors
}

// StartServer starts the peer's server to accept incoming connections
func (p *Peer) StartServer() {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server on %s: %v", addr, err)
	}
	defer listener.Close()
	log.Printf("[INFO] Peer listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[ERROR] Accepting connection: %v", err)
			continue
		}
		go p.handleConnection(conn)
	}
}

// handleConnection processes incoming data
func (p *Peer) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()
		log.Printf("[RECEIVED] %s", data)

		if data == "ready" {
			p.mu.Lock()
			p.ReadyPeers[conn.RemoteAddr().String()] = true
			p.mu.Unlock()
			continue
		}

		parts := strings.Split(data, ",")
		if len(parts) < 3 {
			log.Printf("[ERROR] Invalid message format: %s", data)
			continue
		}

		timestamp, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Printf("[ERROR] Invalid timestamp in message: %s", data)
			continue
		}

		msg := Message{
			Content:   parts[0],
			Timestamp: timestamp,
			Sender:    parts[2],
		}
		p.processMessage(msg)
	}
}

// processMessage adds a received message to the queue and processes it
func (p *Peer) processMessage(msg Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("[INFO] Processing message: %s from %s", msg.Content, msg.Sender)

	p.Clock = max(p.Clock, msg.Timestamp) + 1
	p.MessageQ = append(p.MessageQ, msg)
	p.processQueue()
}

// processQueue processes messages based on their timestamps
func (p *Peer) processQueue() {
	sort.Slice(p.MessageQ, func(i, j int) bool {
		return p.MessageQ[i].Timestamp < p.MessageQ[j].Timestamp
	})

	for len(p.MessageQ) > 0 {
		next := p.MessageQ[0]
		if next.Timestamp > p.Clock+1 {
			break
		}

		fmt.Printf("[CHAT] %s: %s\n", time.Now().Format("15:04:05"), next.Content)
		p.Clock = next.Timestamp
		p.MessageQ = p.MessageQ[1:]
	}
}

// notifyReady informs neighbors the peer is ready
func (p *Peer) notifyReady() {
	time.Sleep(2 * time.Second) // Allow time for all peers to start
	for _, neighbor := range p.Neighbors {
		go func(neighbor string) {
			for {
				conn, err := net.Dial("tcp", neighbor)
				if err != nil {
					log.Printf("[RETRY] Ready notification to %s failed: %v", neighbor, err)
					time.Sleep(1 * time.Second)
					continue
				}
				defer conn.Close()
				fmt.Fprintln(conn, "ready")
				log.Printf("[INFO] Notified neighbor %s that I am ready", neighbor)
				return
			}
		}(neighbor)
	}
}

// waitForNeighborsReady waits for all neighbors to send "ready"
func (p *Peer) waitForNeighborsReady() {
	log.Println("[INFO] Waiting for neighbors to be ready...")
	for {
		p.mu.Lock()
		ready := len(p.ReadyPeers) == len(p.Neighbors)
		p.mu.Unlock()
		if ready {
			log.Println("[INFO] All neighbors are ready.")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// disseminateMessage sends a message to all neighbors
func (p *Peer) disseminateMessage() {
	p.mu.Lock()
	p.Clock++
	timestamp := p.Clock
	p.mu.Unlock()

	content := randomWord()

	for _, neighbor := range p.Neighbors {
		go func(neighbor string) {
			for {
				conn, err := net.Dial("tcp", neighbor)
				if err != nil {
					log.Printf("[RETRY] Connection to neighbor %s failed: %v", neighbor, err)
					time.Sleep(2 * time.Second)
					continue
				}
				defer conn.Close()

				message := fmt.Sprintf("%s,%d,%s", content, timestamp, p.Host)
				fmt.Fprint(conn, message)
				log.Printf("[SENT] Message to %s: %s", neighbor, message)
				return
			}
		}(neighbor)
	}
}

// Main function
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("[ERROR] Usage: %s <host:port> <neighbor1> [neighbor2] ...", os.Args[0])
	}

	hostPort := os.Args[1]
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		log.Fatalf("[ERROR] Invalid host:port format: %s", hostPort)
	}

	host := hostPortParts[0]
	port, err := strconv.Atoi(hostPortParts[1])
	if err != nil {
		log.Fatalf("[ERROR] Invalid port number in host:port: %s", hostPortParts[1])
	}

	neighbors := parseNeighbors(os.Args[2:])
	log.Printf("[INFO] Starting peer on %s:%d with neighbors: %v", host, port, neighbors)

	peer := NewPeer(host, port, neighbors)
	go peer.StartServer()

	peer.notifyReady()
	peer.waitForNeighborsReady()

	pp := NewPoissonProcess(1.0, time.Now().UnixNano()) // 1 message per second
	for {
		interval := pp.TimeForNextEvent()
		time.Sleep(time.Duration(interval * float64(time.Second)))
		peer.disseminateMessage()
	}
}
