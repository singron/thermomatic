package main

import (
	"log"
	"net"

	"github.com/spin-org/thermomatic/internal/server"
)

func main() {
	addr := "[::]:1337"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error listening on %v: %v", addr, err)
	}
	s := server.New(lis)
	if err := s.Serve(); err != nil {
		log.Fatalf("Error in Serve: %v", err)
	}
}
