// client.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	OpUpload   = 1
	OpDownload = 2
	OpAck      = 3
	OpError    = 4
)

type Packet struct {
	Op       byte
	Filename string
	FileSize int64
	Data     []byte
}

func (p *Packet) Encode() ([]byte, error) {
	if len(p.Filename) > 255 {
		return nil, fmt.Errorf("filename too long")
	}
	buf := new(bytes.Buffer)
	if err := buf.WriteByte(p.Op); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(byte(len(p.Filename))); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(p.Filename); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.FileSize); err != nil {
		return nil, err
	}
	if p.Data != nil && len(p.Data) > 0 {
		if _, err := buf.Write(p.Data); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func Decode(r io.Reader) (*Packet, error) {
	p := &Packet{}
	op := make([]byte, 1)
	if _, err := io.ReadFull(r, op); err != nil {
		return nil, err
	}
	p.Op = op[0]

	lenb := make([]byte, 1)
	if _, err := io.ReadFull(r, lenb); err != nil {
		return nil, err
	}
	nameLen := int(lenb[0])

	if nameLen > 0 {
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(r, nameBuf); err != nil {
			return nil, err
		}
		p.Filename = string(nameBuf)
	}

	if err := binary.Read(r, binary.BigEndian, &p.FileSize); err != nil {
		return nil, err
	}

	if p.FileSize > 0 {
		if p.FileSize > (1 << 31) {
			return nil, fmt.Errorf("file too large")
		}
		data := make([]byte, p.FileSize)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		p.Data = data
	} else {
		p.Data = nil
	}
	return p, nil
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage:")
		fmt.Println("  Upload:   go run client.go upload <localfile> <remotefile> <serverAddr:port>")
		fmt.Println("  Download: go run client.go download <remotefile> <localfile> <serverAddr:port>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  go run client.go upload hello.txt server_hello.txt localhost:8080")
		fmt.Println("  go run client.go download server_hello.txt copy.txt localhost:8080")
		return
	}

	action := os.Args[1]
	arg1 := os.Args[2]
	arg2 := os.Args[3]
	serverAddr := os.Args[4]

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	switch action {
	case "upload":
		if err := clientUpload(conn, arg1, arg2); err != nil {
			fmt.Println("Upload failed:", err)
		}
	case "download":
		if err := clientDownload(conn, arg1, arg2); err != nil {
			fmt.Println("Download failed:", err)
		}
	default:
		fmt.Println("Unknown action:", action)
	}
}

func clientUpload(conn net.Conn, localFile, remoteFile string) error {
	f, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()
	if size == 0 {
		// allow empty files, but Data will be nil
	}

	data := make([]byte, size)
	if size > 0 {
		if _, err := io.ReadFull(f, data); err != nil {
			return err
		}
	}

	p := &Packet{
		Op:       OpUpload,
		Filename: remoteFile,
		FileSize: size,
		Data:     data,
	}
	b, err := p.Encode()
	if err != nil {
		return err
	}

	if _, err := conn.Write(b); err != nil {
		return err
	}

	// Wait for ACK or ERROR response
	resp, err := Decode(conn)
	if err != nil {
		return err
	}
	if resp.Op == OpAck {
		fmt.Println("Server ACK:", string(resp.Data))
		return nil
	}
	if resp.Op == OpError {
		return fmt.Errorf("server error: %s", string(resp.Data))
	}
	return fmt.Errorf("unexpected server response op=%d", resp.Op)
}

func clientDownload(conn net.Conn, remoteFile, localFile string) error {
	// send a download request (no data)
	req := &Packet{
		Op:       OpDownload,
		Filename: remoteFile,
		FileSize: 0,
		Data:     nil,
	}
	b, err := req.Encode()
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}

	// server should respond with OpUpload containing file bytes or an error packet
	resp, err := Decode(conn)
	if err != nil {
		return err
	}

	if resp.Op == OpError {
		return fmt.Errorf("server error: %s", string(resp.Data))
	}
	if resp.Op != OpUpload {
		return fmt.Errorf("unexpected server response op=%d", resp.Op)
	}

	// write to local file
	out, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if resp.FileSize > 0 {
		if _, err := out.Write(resp.Data); err != nil {
			return err
		}
	}
	fmt.Println("Downloaded file:", remoteFile, "->", localFile)
	return nil
}
