package main

import (
	"log"
	"net"

	"github.com/spin-org/thermomatic/internal/server"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "[::]:1337")
	if err != nil {
		log.Fatalf("Error resolving listen addr: %v", err)
	}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Error listening on %v: %v", addr, err)
	}
	s := server.New(lis)
	if err := s.Serve(); err != nil {
		log.Fatalf("Error in Serve: %v", err)
	}
}
