package cmd

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"henil.dev/redig/resp"
	"henil.dev/redig/store"
)

type Command = string
type CommandHandler func(conn net.Conn, args []string, kv *store.KVStore) resp.Response

const (
	SetCommand    Command = "set"
	GetCommand    Command = "get"
	PingCommand   Command = "ping"
	DelCommand    Command = "del"
	ExistsCommand Command = "exists"
	IncrCommand   Command = "incr"
	DecrCommand   Command = "decr"
	KeysCommand   Command = "keys"
	ExpireCommand Command = "expire"
	TTLCommand    Command = "ttl"
)

var handlers = map[string]CommandHandler{
	SetCommand:    HandleSetCommand,
	GetCommand:    HandleGetCommand,
	PingCommand:   HandlePingCommand,
	DelCommand:    HandleDelCommand,
	ExistsCommand: HandleExistsCommand,
	IncrCommand:   HandleIncrCommand,
	DecrCommand:   HandleDecrCommand,
	KeysCommand:   HandleKeysCommand,
	ExpireCommand: HandleExpireCommand,
	TTLCommand:    HandleTTLCommand,
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

	var response resp.Response

	if exists {
		response = handler(conn, args, kv)
	} else {
		response = resp.NewError(
			fmt.Sprintf("unknown command '%s'", splitIncoming[0]),
		)
	}

	conn.Write([]byte(response.ToString()))
}

var HandleSetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {

	if len(args) != 2 {
		return resp.NewError(
			"wrong number of arguments for 'set' command",
		)
	}

	key := args[0]
	value := args[1]

	kv.Set(key, value)

	return resp.NewOKResponse()
}

var HandleGetCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'get' command",
		)
	}

	key := args[0]

	value, exists := kv.Get(key)

	response := resp.NewBulkString(value)

	if !exists {
		response.Value = ""
	}

	return response
}

var HandlePingCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) > 1 {
		return resp.NewError(
			"wrong number of arguments for 'ping' command",
		)
	}

	if len(args) == 0 {
		return resp.NewSimpleString("PONG")
	}

	return resp.NewBulkString(args[0])
}

var HandleDelCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'del' command",
		)
	}

	key := args[0]

	didExist := kv.Delete(key)

	return resp.NewIntegerFromBool(didExist)
}

var HandleExistsCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {

	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'exists' command",
		)
	}

	key := args[0]

	exists := kv.Has(key)

	return resp.NewIntegerFromBool(exists)
}

var HandleIncrCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'incr' command",
		)
	}

	key := args[0]

	value, err := kv.Incr(key)

	if err != nil {
		return resp.NewError(
			"value is not an integer or out of range",
		)
	}

	return resp.NewInteger(value)
}

var HandleDecrCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'decr' command",
		)
	}

	key := args[0]

	value, err := kv.Decr(key)

	if err != nil {
		return resp.NewError(
			"value is not an integer or out of range",
		)
	}

	return resp.NewInteger(value)
}

var HandleKeysCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 1 {
		return resp.NewError(
			"wrong number of arguments for 'keys' command",
		)
	}

	pattern := args[0]

	keys := kv.Keys()

	responseSlice := make([]resp.Response, 0, len(keys))

	for _, key := range keys {

		patternMatch, err := filepath.Match(pattern, key)

		if err != nil {
			return resp.NewError("invalid pattern")
		}

		if !patternMatch {
			continue
		}

		responseSlice = append(responseSlice, resp.NewBulkString(key))
	}

	return resp.NewArray(responseSlice)
}

var HandleExpireCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {
	if len(args) != 2 {
		return resp.NewError("wrong number of arguments for 'expire' command")
	}

	key := args[0]
	ttl, err := strconv.Atoi(args[1])

	if err != nil {
		return resp.NewError("value is not an integer or out of range")
	}

	set := kv.Expire(key, ttl)

	return resp.NewIntegerFromBool(set)
}

var HandleTTLCommand CommandHandler = func(conn net.Conn, args []string, kv *store.KVStore) resp.Response {

	if len(args) != 1 {
		return resp.NewError("wrong number of arguments for 'ttl' command")
	}

	key := args[0]
	ttl := kv.TTL(key)

	return resp.NewInteger(ttl)
}
