package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Peer struct {
	Host       string
	Port       int
	RemoteAddr string
	ServerAddr string
	localQueue []string
	mu         sync.Mutex
}

type Token struct {
	HolderID  string
	Timestamp int64
}

func NewPeer(host string, port int, remoteAddr, serverAddr string) *Peer {
	return &Peer{
		Host:       host,
		Port:       port,
		RemoteAddr: remoteAddr,
		ServerAddr: serverAddr,
		localQueue: []string{},
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

// handleConnection processes incoming tokens
func (p *Peer) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Printf("Received: %s", msg)
		if msg == "TOKEN" {
			p.handleToken()
		}
	}
}

// handleToken processes the token, sending requests to the server and forwarding it
func (p *Peer) handleToken() {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Token received. Processing %d requests...", len(p.localQueue))
	for _, request := range p.localQueue {
		p.sendMessageToServer(request)
	}
	p.localQueue = nil

	// Forward the token
	time.Sleep(2 * time.Second) // Simulate processing time
	conn, err := net.Dial("tcp", p.RemoteAddr)
	if err != nil {
		log.Printf("Failed to forward token to %s: %v", p.RemoteAddr, err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "TOKEN\n")
	log.Printf("Token forwarded to %s", p.RemoteAddr)
}

// sendMessageToServer sends a request to the server
func (p *Peer) sendMessageToServer(request string) {
	conn, err := net.Dial("tcp", p.ServerAddr)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()

	// Send the request to the server
	fmt.Fprintf(conn, "%s\n", request)
	log.Printf("Sent request to server: %s", request)

	// Read and log the server's response
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		log.Printf("Received response from server: %s", response)
	} else if err := scanner.Err(); err != nil {
		log.Printf("Error reading server response: %v", err)
	}
}

func main() {
	if len(os.Args) < 5 {
		log.Fatalf("Usage: go run peer.go <host> <port> <remoteAddr> <serverAddr> <startToken>")
	}

	host := os.Args[1]
	port := atoi(os.Args[2])
	remoteAddr := os.Args[3]
	serverAddr := os.Args[4]
	startToken := os.Args[5] == "true"

	peer := NewPeer(host, port, remoteAddr, serverAddr)
	go peer.StartServer()

	if startToken {
		time.Sleep(2 * time.Second) // Wait for other peers to start
		log.Println("Starting the token...")
		conn, err := net.Dial("tcp", peer.RemoteAddr)
		if err != nil {
			log.Fatalf("Failed to send initial token: %v", err)
		}
		defer conn.Close()
		fmt.Fprintf(conn, "TOKEN\n")
	}

	pp := NewPoissonProcess(0.1, time.Now().UnixNano())
	for {
		message := RandomOperation()
		peer.mu.Lock()
		peer.localQueue = append(peer.localQueue, message)
		peer.mu.Unlock()
		time.Sleep(time.Duration(pp.TimeForNextEvent()) * time.Second)
	}
}

// RandomOperation generates a random arithmetic operation
func RandomOperation() string {
	ops := []string{"add", "sub", "mul", "div"}
	op := ops[rand.Intn(len(ops))]
	x := rand.Float64() * 10
	y := rand.Float64() * 10
	return fmt.Sprintf("%s %.2f %.2f", op, x, y)
}

// atoi converts a string to an integer
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Invalid number: %v", err)
	}
	return i
}
