package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"0xgirish.eth/redis/app/resp"
	"0xgirish.eth/redis/app/resp/types"
	"0xgirish.eth/redis/app/store"
)

type CMD = string

const (
	PING CMD = "PING"
	SET  CMD = "SET"
	GET  CMD = "GET"
	DEL  CMD = "DEL"
)

type ConnHandler interface {
	Handle(net.Conn) error
}

type redisConnHandler struct {
	store   *store.Store
	scanner *resp.Scanner
}

// Handle handles a net connection according to RESP protocol
// always run this method in separate go routine
func (r redisConnHandler) Handle(conn net.Conn) error {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	r.scanner = resp.NewScanner(conn)
	for r.scanner.Scan() {
		token := r.scanner.Token()

		fmt.Fprint(conn, r.handle(token).ToResp())
	}

	if err := r.scanner.Err(); err != nil {
		e := types.Error(err.Error())
		fmt.Fprint(conn, e.ToResp())

		return err
	}

	return nil
}

func (r *redisConnHandler) handle(token types.Type) types.Type {
	switch t := token.(type) {
	case *types.Array:
		if t.Len() < 0 {
			err := types.Error("invalid input")
			return &err
		}

		switch tt := t.Index(0).(type) {
		case *types.BulkString:
			switch strings.ToUpper(string(*tt)) {
			case PING:
				return r.ping(t.CastBulkStringFrom(1))
			case SET:
				return r.set(t.CastBulkStringFrom(1))
			case GET:
				return r.get(t.CastBulkStringFrom(1))
			case DEL:
				return r.del(t.CastBulkStringFrom(1))
			}

		}

	}

	return &types.EmptyBulkString
}

func (r *redisConnHandler) ping(args []types.BulkString) types.Type {
	log.Println("handling ping command")

	if len(args) == 0 {
		x := types.String("PONG")
		return &x
	}

	if len(args) > 1 {
		err := types.Error("ERR wrong number of arguments for 'ping' command")
		return &err
	}

	return &args[0]
}

func (r *redisConnHandler) set(args []types.BulkString) types.Type {
	log.Println("handling set command")

	if len(args) < 2 {
		err := types.Error("ERR wrong number of arguments for 'set' command")
		return &err
	}

	r.store.SET(args[0], args[1])
	return &types.OK
}

func (r *redisConnHandler) get(args []types.BulkString) types.Type {
	log.Println("handling get command")

	if len(args) < 1 {
		err := types.Error("ERR wrong number of arguments for 'set' command")
		return &err
	}

	result := r.store.GET(args[0])
	return &result
}

func (r *redisConnHandler) del(args []types.BulkString) types.Type {
	log.Println("handling del command")

	if len(args) < 1 {
		err := types.Error("ERR wrong number of arguments for 'set' command")
		return &err
	}

	result := r.store.DEL(args[0])
	return &result
}
