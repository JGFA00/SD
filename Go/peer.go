package main

import (
    "bufio"
    "fmt"
    "log"
    "math/rand"
    "net"
    "os"
    "strconv"
    "strings"
    "time"
)

// Peer struct holds information about the peer's host and port
type Peer struct {
    Host       string
    Port       int
    RemoteAddr string
}

// NewPeer creates a new Peer with the given host, port, and remote address
func NewPeer(host string, port int, remoteAddr string) *Peer {
    return &Peer{
        Host:       host,
        Port:       port,
        RemoteAddr: remoteAddr,
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

// handleConnection processes an incoming connection, reads messages, and sends responses
func (p *Peer) handleConnection(conn net.Conn) {
    defer conn.Close()
    clientAddr := conn.RemoteAddr().String()
    log.Printf("Connected to client: %s", clientAddr)

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        msg := scanner.Text()
        log.Printf("Received from %s: %s", clientAddr, msg)

        // Parse the message as a calculator command
        result, err := p.calculate(msg)
        if err != nil {
            _, _ = conn.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
            continue
        }

        // Send result back to client
        _, err = conn.Write([]byte(fmt.Sprintf("Result: %f\n", result)))
        if err != nil {
            log.Printf("Failed to send response to %s: %v", clientAddr, err)
            return
        }
    }
}

// SendMessage sends a message to the remote peer
func (p *Peer) SendMessage(message string) {
    conn, err := net.Dial("tcp", p.RemoteAddr)
    if err != nil {
        log.Printf("Error connecting to peer at %s: %v", p.RemoteAddr, err)
        return
    }
    defer conn.Close()

    fmt.Fprintf(conn, "%s\n", message)
    response, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        log.Printf("Error reading response from peer: %v", err)
        return
    }
    log.Printf("Received response: %s", strings.TrimSpace(response))
}

// RandomOperation generates a random arithmetic operation with random operands
func RandomOperation() string {
    operations := []string{"add", "sub", "mul", "div"}
    op := operations[rand.Intn(len(operations))]
    num1 := rand.Float64() * 10
    num2 := rand.Float64() * 10
    return fmt.Sprintf("%s %.2f %.2f", op, num1, num2)
}

// calculate parses and performs the arithmetic operation
func (p *Peer) calculate(input string) (float64, error) {
    parts := strings.Fields(input)
    if len(parts) != 3 {
        return 0, fmt.Errorf("invalid format; expected '<operation> <num1> <num2>'")
    }

    op := parts[0]
    num1, err1 := strconv.ParseFloat(parts[1], 64)
    num2, err2 := strconv.ParseFloat(parts[2], 64)

    if err1 != nil || err2 != nil {
        return 0, fmt.Errorf("invalid numbers")
    }

    switch op {
    case "add":
        return num1 + num2, nil
    case "sub":
        return num1 - num2, nil
    case "mul":
        return num1 * num2, nil
    case "div":
        if num2 == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return num1 / num2, nil
    default:
        return 0, fmt.Errorf("unsupported operation: %s", op)
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

// main function initializes the peer and periodically sends random messages to the remote peer
func main() {
    if len(os.Args) < 4 {
        log.Fatalf("Usage: go run peer.go <host> <port> <remoteAddr>")
    }

    host := os.Args[1]
    port := os.Args[2]
    remoteAddr := os.Args[3]

    peer := NewPeer(host, atoi(port), remoteAddr)
    pp := NewPoissonProcess(0.5, time.Now().UnixNano()) // Adjust lambda as needed

    // Start the server in a separate goroutine
    go peer.StartServer()

    // Continuously send random operations based on Poisson intervals
    for {
        message := RandomOperation()
        log.Printf("Sending message: %s", message)
        peer.SendMessage(message)

        // Wait for the next event based on Poisson process interval
        time.Sleep(time.Duration(pp.TimeForNextEvent()) * time.Second)
    }
}


