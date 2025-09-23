// server.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const (
	OpUpload   = 1 // used for sending file bytes (both directions)
	OpDownload = 2 // request to download a file (client -> server)
	OpAck      = 3 // acknowledgement
	OpError    = 4 // error with message in Data
)

type Packet struct {
	Op       byte
	Filename string
	FileSize int64
	Data     []byte
}

// Encode serializes a Packet to bytes
func (p *Packet) Encode() ([]byte, error) {
	if len(p.Filename) > 255 {
		return nil, fmt.Errorf("filename too long")
	}
	buf := new(bytes.Buffer)

	// OP
	if err := buf.WriteByte(p.Op); err != nil {
		return nil, err
	}

	// filename length + filename
	if err := buf.WriteByte(byte(len(p.Filename))); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(p.Filename); err != nil {
		return nil, err
	}

	// file size (8 bytes)
	if err := binary.Write(buf, binary.BigEndian, p.FileSize); err != nil {
		return nil, err
	}

	// data (if any)
	if p.Data != nil && len(p.Data) > 0 {
		if _, err := buf.Write(p.Data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Decode reads bytes from r to construct a Packet.
// It reads header fields first, then reads FileSize bytes into Data if needed.
func Decode(r io.Reader) (*Packet, error) {
	p := &Packet{}

	// 1 byte OP
	op := make([]byte, 1)
	if _, err := io.ReadFull(r, op); err != nil {
		return nil, err
	}
	p.Op = op[0]

	// 1 byte filename length
	lenb := make([]byte, 1)
	if _, err := io.ReadFull(r, lenb); err != nil {
		return nil, err
	}
	nameLen := int(lenb[0])

	// filename
	if nameLen > 0 {
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(r, nameBuf); err != nil {
			return nil, err
		}
		p.Filename = string(nameBuf)
	} else {
		p.Filename = ""
	}

	// file size (8 bytes)
	if err := binary.Read(r, binary.BigEndian, &p.FileSize); err != nil {
		return nil, err
	}

	// If file size > 0, read Data
	if p.FileSize > 0 {
		if p.FileSize > (1 << 31) { // safety: 2GB limit in this simple example
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
	const listenAddr = ":8080"

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()
	fmt.Println("File server listening on", listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected:", conn.RemoteAddr())

	// We'll repeatedly decode packets from the same connection.
	// Note: Decode uses io.ReadFull, which blocks until expected bytes are read.
	for {
		pkt, err := Decode(conn)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Decode error:", err)
			}
			return
		}

		switch pkt.Op {
		case OpUpload:
			if err := handleUpload(pkt); err != nil {
				sendError(conn, err.Error())
				fmt.Println("Upload error:", err)
			} else {
				sendAck(conn, "upload successful")
				fmt.Println("Saved file:", pkt.Filename)
			}

		case OpDownload:
			if err := handleDownload(conn, pkt.Filename); err != nil {
				sendError(conn, err.Error())
				fmt.Println("Download error:", err)
			} else {
				fmt.Println("Served file:", pkt.Filename)
			}

		default:
			sendError(conn, "unknown operation")
		}
	}
}

func handleUpload(p *Packet) error {
	// ensure directory exists
	dir := filepath.Dir(p.Filename)
	if dir == "." || dir == "" {
		// current directory — ok
	} else {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(p.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if p.Data != nil && len(p.Data) > 0 {
		if _, err := f.Write(p.Data); err != nil {
			return err
		}
	}
	return nil
}

func handleDownload(conn net.Conn, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()
	data := make([]byte, size)
	if _, err := io.ReadFull(f, data); err != nil {
		return err
	}

	// build an upload-style packet to send the file
	resp := &Packet{
		Op:       OpUpload,
		Filename: filename,
		FileSize: size,
		Data:     data,
	}
	bytesOut, err := resp.Encode()
	if err != nil {
		return err
	}
	_, err = conn.Write(bytesOut)
	return err
}

func sendAck(conn net.Conn, msg string) {
	p := &Packet{
		Op:       OpAck,
		Filename: "",
		FileSize: int64(len(msg)),
		Data:     []byte(msg),
	}
	b, _ := p.Encode()
	conn.Write(b)
}

func sendError(conn net.Conn, msg string) {
	p := &Packet{
		Op:       OpError,
		Filename: "",
		FileSize: int64(len(msg)),
		Data:     []byte(msg),
	}
	b, _ := p.Encode()
	conn.Write(b)
}
