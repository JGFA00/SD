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
    "sync"
    "time"
)

// Peer struct holds information about the peer's host, port, and the next peer's address
type Peer struct {
    Host       string
    Port       int
    RemoteAddr string
    localQueue []string
    mu         sync.Mutex // Protects access to localQueue
}

// Token struct represents the token used in the token ring
// It may contain additional information, such as the holder ID or a timestamp
type Token struct {
    HolderID   string // ID of the peer currently holding the token
    Timestamp  int64  // Time the token was last forwarded (optional)
}

// NewPeer creates a new Peer with the given host, port, and remote address
func NewPeer(host string, port int, remoteAddr string) *Peer {
    return &Peer{
        Host:       host,
        Port:       port,
        RemoteAddr: remoteAddr,
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

// handleConnection processes an incoming connection, reads messages, and sends responses
func (p *Peer) handleConnection(conn net.Conn) {
    defer conn.Close()
    clientAddr := conn.RemoteAddr().String()
    log.Printf("Connected to client: %s", clientAddr)

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        msg := scanner.Text()
        log.Printf("Received from %s: %s", clientAddr, msg)

        if msg == "TOKEN" {
            p.handleToken(Token{HolderID: p.Host, Timestamp: time.Now().Unix()})
            continue
        }

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

// handleToken processes the received token
func (p *Peer) handleToken(token Token) {
    p.mu.Lock()
    defer p.mu.Unlock()

    log.Printf("Token received by %s with %d requests in queue", p.Host, len(p.localQueue))
    if len(p.localQueue) > 0 {
        // Process each request in the queue
        for _, req := range p.localQueue {
            // Send request to the server or handle it locally as needed
            p.sendMessageToServer(req)
        }
        // Clear the queue after processing
        p.localQueue = []string{}
    }
    // Add delay before forwarding the token
    time.Sleep(15 * time.Second) // 15-second delay between token exchanges
    p.forwardToken()
}

// forwardToken forwards the token to the next peer in the ring.
// If it fails to connect to the next peer, it retries and processes local requests.
func (p *Peer) forwardToken() {
    token := Token{HolderID: p.Host, Timestamp: time.Now().Unix()}
    retryLimit := 5           // Maximum number of retries
    retryInterval := 5 * time.Second // Interval between retries

    for retries := 0; retries < retryLimit; retries++ {
        conn, err := net.Dial("tcp", p.RemoteAddr)
        if err != nil {
            log.Printf("Failed to connect to next peer (%s): %v", p.RemoteAddr, err)

            // Process local requests while retrying
            p.processLocalQueue()

            // Wait before retrying
            time.Sleep(retryInterval)
            continue
        }
        defer conn.Close()

        // Send the token to the next peer
        fmt.Fprintf(conn, "TOKEN\n")
        log.Printf("Token (HolderID: %s) forwarded to %s", token.HolderID, p.RemoteAddr)
        return // Token successfully forwarded
    }

    // If all retries fail, process the queue locally as fallback
    log.Printf("Max retries reached. Processing requests locally.")
    p.processLocalQueue()
}

// SendMessage sends a message to the remote peer
func (p *Peer) QueueMessage(message string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.localQueue = append(p.localQueue, message)
    log.Printf("Message added to queue: %s", message)
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

// sendMessageToServer sends a message to the server for processing
func (p *Peer) sendMessageToServer(message string) {
        // Replace with the actual server address
    serverAddr := "localhost:8080" // Ensure this is the actual server address
    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        log.Printf("Failed to connect to server: %v", err)
        return
    }
    defer conn.Close()

    // Send the message
    fmt.Fprintf(conn, "%s\n", message)
    log.Printf("Sent message to server: %s", message)

    // Read the server's response
    scanner := bufio.NewScanner(conn)
    if scanner.Scan() {
        response := scanner.Text()
        log.Printf("Received response from server: %s", response)
    } else if err := scanner.Err(); err != nil {
        log.Printf("Error reading response from server: %v", err)
    }
}



// processLocalQueue processes requests in the local queue by sending them to the server.
func (p *Peer) processLocalQueue() {
    p.mu.Lock()
    defer p.mu.Unlock()

    for len(p.localQueue) > 0 {
        // Retrieve the next request from the queue
        request := p.localQueue[0]
        p.localQueue = p.localQueue[1:] // Remove the request from the queue

        log.Printf("Processing request locally: %s", request)
        p.sendMessageToServer(request)
    }
}


// main function initializes the peer and periodically sends random messages to the remote peer
func main() {
    if len(os.Args) < 5 {
        log.Fatalf("Usage: go run peer.go <host> <port> <remoteAddr> <startToken>")
    }

    host := os.Args[1]
    port := os.Args[2]
    remoteAddr := os.Args[3]
    startToken := os.Args[4] == "true"

    peer := NewPeer(host, atoi(port), remoteAddr)
    pp := NewPoissonProcess(0.1, time.Now().UnixNano())

    // Start the server in a separate goroutine
    go peer.StartServer()

    // If startToken is true, initiate the token rotation
    if startToken {
        log.Printf("Starting token...")
        peer.forwardToken()
    }

    // Continuously generate random operations based on Poisson intervals and add to queue
    for {
        message := RandomOperation()
        peer.QueueMessage(message)

        // Wait for the next event based on Poisson process interval
        time.Sleep(time.Duration(pp.TimeForNextEvent()) * time.Second)
    }
}
