package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage:")
		fmt.Println("UPLOAD: go run client.go upload <localfile> <remotefile>")
		fmt.Println("DOWNLOAD: go run client.go download <remotefile> <localfile>")
		return
	}

	action, source, target := os.Args[1], os.Args[2], os.Args[3]
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	defer conn.Close()

	if action == "upload" {
		uploadFile(conn, source, target)
	} else if action == "download" {
		downloadFile(conn, source, target)
	} else {
		fmt.Println("Unkown actions:", action)
	}
}

func uploadFile(conn net.Conn, localFile, remoteFile string) {
	// Tell the server what we're doing
	fmt.Fprintf(conn, "UPLOAD %s\n", remoteFile)

	file, err := os.Open(localFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Send the file contents
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return
	}
	fmt.Println("File uploaded:", localFile, "→", remoteFile)
}

func downloadFile(conn net.Conn, remoteFile, localFile string) {
	// Tell the server what we're doing
	fmt.Fprintf(conn, "DOWNLOAD %s\n", remoteFile)

	file, err := os.Create(localFile)
	if err != nil {
		fmt.Println("Error creating local file:", err)
		return
	}
	defer file.Close()

	// Receive file contents
	_, err = io.Copy(file, conn)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	fmt.Println("File downloaded:", remoteFile, "→", localFile)
}
