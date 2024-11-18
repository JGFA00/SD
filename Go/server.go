package main

import (
    "bufio"
    "fmt"
    "net"
    "strconv"
    "strings"
)

// Handle incoming connections
func handleConnection(conn net.Conn) {
    defer conn.Close()
    fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        command := scanner.Text()
        fmt.Printf("Received command: %s\n", command)

        // Process command
        parts := strings.Fields(command) // Split by whitespace
        if len(parts) != 3 {
            fmt.Fprintf(conn, "Invalid format. Use <operation> <x> <y>\n")
            continue
        }

        op := parts[0]
        x, err1 := strconv.ParseFloat(parts[1], 64)
        y, err2 := strconv.ParseFloat(parts[2], 64)

        if err1 != nil || err2 != nil {
            fmt.Fprintf(conn, "Invalid numbers provided.\n")
            continue
        }

        var result float64
        var resultMsg string
        switch op {
        case "add":
            result = x + y
            resultMsg = fmt.Sprintf("Result: %f\n", result)
        case "sub":
            result = x - y
            resultMsg = fmt.Sprintf("Result: %f\n", result)
        case "mul":
            result = x * y
            resultMsg = fmt.Sprintf("Result: %f\n", result)
        case "div":
            if y == 0 {
                resultMsg = "Error: Division by zero\n"
            } else {
                result = x / y
                resultMsg = fmt.Sprintf("Result: %f\n", result)
            }
        default:
            resultMsg = "Invalid operation. Supported operations: add, sub, mul, div\n"
        }

        // Send result back to client
        fmt.Fprintf(conn, resultMsg)
    }
}

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()
    fmt.Println("Server is listening on port 8080")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleConnection(conn) // Handle each connection concurrently
    }
}
