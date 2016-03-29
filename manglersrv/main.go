/* libmangler server is naught but a chatbot at the moment */
package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const (
	protoVersion = 2
	protoPort    = 40000
)

// handle handles a network connection and chats with the client
func handle(conn net.Conn) {
	defer conn.Close()
	defer log.Println("client", conn.RemoteAddr(), "disconnected")

handle:
	for {
		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("conn.Read error:", err)
				return
			} else if err == io.EOF && n == 0 {
				return
			}
		}

		// Empty message?
		if n == 0 {
			continue
		}

		for i := 0; i < len(buf); {
			switch buf[i] {
			case ' ': // XXX do general whitespace check
				i++
			case 'h':
				fmt.Fprintln(conn, "Be welcomed, friend of libmangler!")
				continue handle
			case 'q':
				fmt.Fprintln(conn, "Farewell, milord!")
				return
			case '\n':
				continue handle
			default:
				fmt.Fprintln(conn, "What?")
				continue handle
			}
		}
	}
}

func main() {
	log.Println("libmangler proto", protoVersion)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", protoPort))
	if err != nil {
		log.Panicln("net.Listen failed:", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("ln.Accept failed:", err)
			continue
		}

		go handle(conn)
	}
}
