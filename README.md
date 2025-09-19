# File Transfer Server and Client

This project is a **case study** for building a Distributed File System (DFS).  
It demonstrates the foundations of networking and file transfer between a client and server using **Go** and **TCP sockets**.  

The goal of this project is to practice handling connections, defining a simple communication protocol, and transferring files reliably — all of which are stepping stones toward a more advanced DFS.

---

## Features

- Simple TCP server that listens for client connections
- Supports two commands:
  - `UPLOAD <filename>` → client sends a file to the server
  - `DOWNLOAD <filename>` → client requests a file from the server
- Client program to interact with the server
- Basic protocol using command headers followed by raw file data

---

## Getting Started

### 1. Clone the repository
```bash
git clone https://github.com/yourusername/File-Transfer.git
cd File-Transfer
```

2. Start the server
```bash
go run server.go
```
You should see:
```bash
File server listening on port 8080...
```
3. Upload a file
First, create a test file:
```bash
echo "Hello this is an example file!" > hello.txt
```
Then run the client:
```bash
go run client.go upload hello.txt server_hello.txt
```
hello.txt → local file on your machine
server_hello.txt → name to store the file on the server

4. Download a file
```bash
go run client.go download server_hello.txt local_copy.txt
```

server_hello.txt → file stored on the server
local_copy.txt → name for the file on your machine

Example Workflow
Server output:
```
File server listening on port 8080...
File received server_hello.txt
File sent: server_hello.txt
```
Client output (upload):
```
File uploaded: hello.txt → server_hello.txt
```
Client output (download):
```
File downloaded: server_hello.txt → local_copy.txt
```
Next Steps

This project is intentionally simple. Future improvements will make the system more robust and closer to a true DFS, such as:

- Adding file size headers for precise transfers
- Supporting multiple requests over a single connection
- Error handling and retries
- Authentication and access control
- Replication across multiple nodes


