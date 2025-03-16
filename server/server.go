package server

import (
	"io"
	"log"
	"net"

	"github.com/henilmalaviya/redig/cmd"
	"github.com/henilmalaviya/redig/store"
)

func NewTCPListener() (*net.Listener, error) {
	listener, err := net.Listen("tcp", ":4001")

	if err != nil {
		return nil, err
	}

	log.Println("Listening on TCP server")

	return &listener, nil
}

func ListenAndAcceptIncomingConnections(listener *net.Listener, kv *store.KVStore) {
	for {
		conn, err := (*listener).Accept()

		if err != nil {
			log.Println("Error accepting TCP connection")
			continue
		}

		log.Printf("Connection accepted from %s\n", conn.RemoteAddr().String())

		go handleConnection(conn, kv)
	}
}

func handleConnection(conn net.Conn, kv *store.KVStore) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		len, err := conn.Read(buffer)

		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed from %s\n", conn.RemoteAddr().String())
				break
			}

			log.Printf("Error reading from TCP connection: %s\n", err.Error())
			break
		}

		go cmd.HandleMessage(conn, string(buffer[:len]), kv)

	}

}
