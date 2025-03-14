package cmd

import (
	"fmt"
	"log"
	"net"
	"strings"

	"henil.dev/redig/resp"
	"henil.dev/redig/store"
)

type Command = string
type CommandHandler func(conn net.Conn, args []string, kv *store.KVStore)

const (
	SetCommand    Command = "set"
	GetCommand    Command = "get"
	PingCommand   Command = "ping"
	DelCommand    Command = "del"
	ExistsCommand Command = "exists"
)

var handlers = map[string]CommandHandler{
	SetCommand:    HandleSetCommand,
	GetCommand:    HandleGetCommand,
	PingCommand:   HandlePingCommand,
	DelCommand:    HandleDelCommand,
	ExistsCommand: HandleExistsCommand,
}

func HandleMessage(conn net.Conn, incoming string, kv *store.KVStore) {
	log.Printf("Message received: %s\n", incoming)

	strippedIncoming := strings.TrimSpace(incoming)

	if strippedIncoming == "" {
		return
	}

	splitIncoming := strings.Split(strippedIncoming, " ")

	log.Printf("Split incoming: %v\n", splitIncoming)

	rootCommand, args := splitIncoming[0], splitIncoming[1:]

	rootCommand = strings.ToLower(rootCommand)

	handler, exists := handlers[rootCommand]

	if exists {
		handler(conn, args, kv)
		return
	}

	conn.Write(
		resp.NewErrorResponse(
			fmt.Sprintf("unknown command '%s'", splitIncoming[0]),
		).Bytes(),
	)
}

var HandleSetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {

	if len(args) != 2 {
		conn.Write(
			resp.NewErrorResponse(
				"wrong number of arguments for 'set' command",
			).Bytes(),
		)
		return
	}

	key := args[0]
	value := args[1]

	kv.Set(key, value)

	conn.Write(
		resp.NewOKResponse().Bytes(),
	)
}

var HandleGetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write(
			resp.NewErrorResponse(
				"wrong number of arguments for 'get' command",
			).Bytes(),
		)
		return
	}

	key := args[0]

	value, exists := kv.Get(key)

	response := resp.NewResponse(resp.BulkStringType, value)

	if !exists {
		response.Value = ""
	}

	conn.Write(response.Bytes())
}

var HandlePingCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) > 1 {
		conn.Write(
			resp.NewErrorResponse("wrong number of arguments for 'ping' command").Bytes(),
		)
		return
	}

	if len(args) == 0 {
		conn.Write(resp.NewResponse(resp.SimpleStringType, "PONG").Bytes())
		return
	}

	conn.Write(resp.NewResponse(resp.BulkStringType, args[0]).Bytes())
}

var HandleDelCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write(
			resp.NewErrorResponse(
				"wrong number of arguments for 'del' command",
			).Bytes(),
		)
		return
	}

	key := args[0]

	didExist := kv.Delete(key)

	conn.Write(resp.NewIntegerResponseFromBool(didExist).Bytes())
}

var HandleExistsCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {

	if len(args) != 1 {
		conn.Write(
			resp.NewErrorResponse(
				"wrong number of arguments for 'exists' command",
			).Bytes(),
		)
		return
	}

	key := args[0]

	exists := kv.Has(key)

	conn.Write(resp.NewIntegerResponseFromBool(exists).Bytes())
}
