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
	IncrCommand   Command = "incr"
	DecrCommand   Command = "decr"
)

var handlers = map[string]CommandHandler{
	SetCommand:    HandleSetCommand,
	GetCommand:    HandleGetCommand,
	PingCommand:   HandlePingCommand,
	DelCommand:    HandleDelCommand,
	ExistsCommand: HandleExistsCommand,
	IncrCommand:   HandleIncrCommand,
	DecrCommand:   HandleDecrCommand,
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
		[]byte(resp.NewError(
			fmt.Sprintf("unknown command '%s'", splitIncoming[0]),
		).ToString()),
	)
}

var HandleSetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {

	if len(args) != 2 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'set' command",
		).ToString()))
		return
	}

	key := args[0]
	value := args[1]

	kv.Set(key, value)

	conn.Write([]byte(resp.NewOKResponse().ToString()))
}

var HandleGetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'get' command",
		).ToString()))
		return
	}

	key := args[0]

	value, exists := kv.Get(key)

	response := resp.NewBulkString(value)

	if !exists {
		response.Value = ""
	}

	conn.Write([]byte(response.ToString()))
}

var HandlePingCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) > 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'ping' command",
		).ToString()))
		return
	}

	if len(args) == 0 {
		conn.Write([]byte(resp.NewSimpleString("PONG").ToString()))
		return
	}

	conn.Write([]byte(resp.NewBulkString(args[0]).ToString()))
}

var HandleDelCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'del' command",
		).ToString()))
		return
	}

	key := args[0]

	didExist := kv.Delete(key)

	conn.Write([]byte(resp.NewIntegerFromBool(didExist).ToString()))
}

var HandleExistsCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {

	if len(args) != 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'exists' command",
		).ToString()))
		return
	}

	key := args[0]

	exists := kv.Has(key)

	conn.Write([]byte(resp.NewIntegerFromBool(exists).ToString()))
}

var HandleIncrCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'incr' command",
		).ToString()))
		return
	}

	key := args[0]

	value, err := kv.Incr(key)

	if err != nil {
		conn.Write([]byte(resp.NewError(
			"value is not an integer or out of range",
		).ToString()))
		return
	}

	conn.Write([]byte(resp.NewInteger(value).ToString()))
}

var HandleDecrCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) {
	if len(args) != 1 {
		conn.Write([]byte(resp.NewError(
			"wrong number of arguments for 'decr' command",
		).ToString()))
		return
	}

	key := args[0]

	value, err := kv.Decr(key)

	if err != nil {
		conn.Write([]byte(resp.NewError(
			"value is not an integer or out of range",
		).ToString()))
		return
	}

	conn.Write([]byte(resp.NewInteger(value).ToString()))
}
