package cmd

import (
	"fmt"
	"log"
	"net"
	"strings"

	"henil.dev/redig/resp"
	"henil.dev/redig/store"
)

func HandleSetCommand(conn net.Conn, splitIncoming []string, kv *store.KVStore) {

	if len(splitIncoming) != 3 {
		conn.Write(
			resp.NewResponse(
				resp.ErrorType,
				"wrong number of arguments for 'set' command",
			).Bytes(),
		)
		return
	}

	key := splitIncoming[1]
	value := splitIncoming[2]

	kv.Set(key, value)

	conn.Write(
		resp.NewResponse(
			resp.SimpleStringType,
			"OK",
		).Bytes(),
	)
}

func HandleGetCommand(conn net.Conn, splitIncoming []string, kv *store.KVStore) {
	if len(splitIncoming) != 2 {
		conn.Write(
			resp.NewResponse(
				resp.ErrorType,
				"wrong number of arguments for 'get' command",
			).Bytes(),
		)
		return
	}

	key := splitIncoming[1]

	value, _ := kv.Get(key)

	response := resp.NewResponse(resp.BulkStringType, value)

	conn.Write(response.Bytes())
}

func HandleMessage(conn net.Conn, incoming string, kv *store.KVStore) {
	log.Printf("Message received: %s\n", incoming)

	strippedIncoming := strings.TrimSpace(incoming)

	if strippedIncoming == "" {
		return
	}

	splitIncoming := strings.Split(strippedIncoming, " ")

	log.Printf("Split incoming: %v\n", splitIncoming)

	if strings.EqualFold(splitIncoming[0], "SET") {
		HandleSetCommand(conn, splitIncoming, kv)
		return
	}

	if strings.EqualFold(splitIncoming[0], "GET") {
		HandleGetCommand(conn, splitIncoming, kv)
		return
	}

	conn.Write(
		resp.NewResponse(
			resp.ErrorType,
			fmt.Sprintf("unknown command '%s'", splitIncoming[0]),
		).Bytes(),
	)
}
