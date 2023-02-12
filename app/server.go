package main

import (
	"fmt"
	"net"
	"os"

	"0xgirish.eth/redis/app/store"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	var h ConnHandler = redisConnHandler{
		store: store.New(),
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(fmt.Errorf("Error accepting connection: %w", err))
		}

		go h.Handle(conn)
	}
}
