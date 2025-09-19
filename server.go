package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("File server listenning on port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting conneciton:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		commandLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}

		commandLine = strings.TrimSpace(commandLine)
		parts := strings.Split(commandLine, " ")
		if len(parts) < 2 {
			conn.Write([]byte("Invalid command\n"))
			continue
		}

		cmd, filename := parts[0], parts[1]

		if cmd == "UPLOAD" {
			receiveFile(conn, reader, filename)
		} else if cmd == "DOWNLOAD" {
			sendFile(conn, filename)
		} else {
			conn.Write([]byte("Uknown command\n"))
		}

	}
}

func receiveFile(conn net.Conn, reader *bufio.Reader, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		conn.Write([]byte("Error creating file\n"))
		return
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		fmt.Println("Error receiving file:", err)
		return
	}
	fmt.Println("File received", filename)
}

func sendFile(conn net.Conn, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		conn.Write([]byte("Error Opening file\n"))
		return
	}
	defer file.Close()

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return
	}
	fmt.Println("File sent:", filename)

}
