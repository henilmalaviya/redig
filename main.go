package main

import (
	"log"

	"henil.dev/redig/server"
	"henil.dev/redig/store"
)

func main() {
	var kv = store.NewKVStore()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	listener, err := server.NewTCPListener()

	if err != nil {
		log.Fatalf("Failed to create TCP listener: %s\n", err.Error())
	}

	defer (*listener).Close()

	server.ListenAndAcceptIncomingConnections(listener, kv)

}
