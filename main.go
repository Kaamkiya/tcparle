package main

import (
	"flag"
	"net"
	"log"
	"strconv"

	"codeberg.org/Kaamkiya/tcparle/internal/chat"
)

func main() {
	flagPort := flag.Int("port", 8888, "The port on which to run the server.")
	flag.Parse()

	srv := chat.NewServer()
	go srv.RunCommands()

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(*flagPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error: failed to accept connection: %s\n", err.Error())
			continue
		}

		go srv.NewClient(conn)
	}
}
