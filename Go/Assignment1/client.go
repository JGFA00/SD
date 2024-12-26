package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run client.go <server IP> <server Port>")
        return
    }

    serverAddr := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return
    }
    defer conn.Close()
    fmt.Printf("Connected to server: %s\n", serverAddr)

    reader := bufio.NewReader(os.Stdin)
    scanner := bufio.NewScanner(conn)
    for {
        fmt.Print("$ ")
        command, _ := reader.ReadString('\n')
        command = strings.TrimSpace(command)

        if command == "quit" {
            break
        }

        // Send command to server
        fmt.Fprintln(conn, command)

        // Receive and print response
        if scanner.Scan() {
            fmt.Printf("Result: %s\n", scanner.Text())
        }
    }
}
