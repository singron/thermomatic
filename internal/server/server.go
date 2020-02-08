// Package server defines a TCP server that accepts readings from incoming
// connections and prints them to stdout.
package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
	"github.com/spin-org/thermomatic/internal/imei"
)

// Server will accept and handle connections.
type Server struct {
	lis net.Listener

	// Override these for testing

	output  io.Writer
	timeNow func() time.Time
}

// New returns a Server that accepts from lis and writes to os.Stdout.
func New(lis net.Listener) *Server {
	return &Server{lis: lis, output: os.Stdout, timeNow: time.Now}
}

// Serve accepts and handles connections. It only returns if accept returns an
// error.
func (s *Server) Serve() error {
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			return fmt.Errorf("server: accept: %w", err)
		}
		go s.handleConn(conn)
	}
}

func readFullBuf(conn net.Conn, buf []byte) error {
	n := 0
	for n < len(buf) {
		c, err := conn.Read(buf[n:])
		if err != nil {
			return err
		}
		n += c
	}
	return nil
}

// handleConn will handle conn until there is an error. handleConn will close
// the connection before returning.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted client[%v]", conn.RemoteAddr())
	var loginBuf [15]byte
	conn.SetReadDeadline(time.Now().Add(time.Second))
	if err := readFullBuf(conn, loginBuf[:]); err != nil {
		log.Printf("Error reading client[%v] login: %v", conn.RemoteAddr(), err)
		return
	}
	code, err := imei.Decode(loginBuf[:])
	if err != nil {
		log.Printf("Error decoding client[%v] login: %v", conn.RemoteAddr(), err)
		return
	}

	var out bytes.Buffer
	for {
		var r client.Reading
		var readingBuf [40]byte
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err := readFullBuf(conn, readingBuf[:]); err != nil {
			if err == io.EOF {
				// Probably a normal closure.
				log.Printf("Connection closed for client[%v]", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading client[%v] reading: %v", conn.RemoteAddr(), err)
			return
		}
		timestamp := s.timeNow()
		if ok := r.Decode(readingBuf[:]); !ok {
			log.Printf("Error decoding client[%v] reading", conn.RemoteAddr())
			return
		}
		fmt.Fprintf(&out, "%d,%d,%g,%g,%g,%g,%g\n", timestamp.UnixNano(), code,
			r.Temperature, r.Altitude, r.Latitude, r.Longitude, r.BatteryLevel)
		s.output.Write(out.Bytes())
		out.Reset()
	}
}
