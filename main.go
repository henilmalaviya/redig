package main

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type KVStore struct {
	store map[string]string
	mutex sync.RWMutex
}

func NewKVStore() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (s *KVStore) Set(key string, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[key] = value
}

func (s *KVStore) Has(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.store[key]
	return exists
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.store[key]
	return value, exists
}

func (s *KVStore) Delete(key string) bool {
	if s.Has(key) {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		delete(s.store, key)
		return true
	}

	return false
}

var kv = NewKVStore()

func handleSetCommand(conn net.Conn, splitIncoming []string) {

	if len(splitIncoming) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
		return
	}

	key := splitIncoming[1]
	value := splitIncoming[2]

	kv.Set(key, value)

	conn.Write([]byte("+OK\r\n"))
}

func handleGetCommand(conn net.Conn, splitIncoming []string) {
	if len(splitIncoming) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
		return
	}

	key := splitIncoming[1]

	value, exists := kv.Get(key)

	if exists {
		conn.Write([]byte("$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"))
		return
	}

	conn.Write([]byte("$-1\r\n"))
}

func handleMessage(conn net.Conn, incoming string) {
	log.Printf("Message received: %s\n", incoming)

	strippedIncoming := strings.TrimSpace(incoming)

	if strippedIncoming == "" {
		return
	}

	splitIncoming := strings.Split(strippedIncoming, " ")

	log.Printf("Split incoming: %v\n", splitIncoming)

	if strings.EqualFold(splitIncoming[0], "SET") {
		handleSetCommand(conn, splitIncoming)
		return
	}

	if strings.EqualFold(splitIncoming[0], "GET") {
		handleGetCommand(conn, splitIncoming)
		return
	}

}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)

	for {
		len, err := conn.Read(buffer)

		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed from %s\n", conn.RemoteAddr().String())
				break
			}

			log.Fatalf("Error reading from TCP connection: %s\n", err.Error())
			continue
		}

		go handleMessage(conn, string(buffer[:len]))

	}

}

func main() {

	listener, err := net.Listen("tcp", ":4001")

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if err != nil {
		panic("Error listening TCP server")
	}

	log.Println("Listening on TCP server")
	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatalln("Error accepting TCP connection")
			continue
		}

		log.Printf("Connection accepted from %s\n", conn.RemoteAddr().String())

		go handleConnection(conn)
	}
}
