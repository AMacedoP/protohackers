package main

import (
	"io"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	log.Printf("listening on %s", listener.Addr().String())
	if err != nil {
		log.Fatalf("failed to open TCP listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %v", err)
		}

		go func() {
			defer func() {
				log.Println("closing connection")
				if err := conn.Close(); err != nil {
					log.Printf("error closing the connection: %v", err)
				}
			}()
			log.Printf("handling connection for: %s", conn.RemoteAddr().String())
			written, err := io.Copy(conn, conn)
			if err != nil {
				log.Printf("failed in io.Copy %v", err)
			}
			log.Printf("written %d bytes", written)
		}()
	}
}
