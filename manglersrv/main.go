/* Manglersrv implements the v and q simple commands at the moment. */
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"unicode"
)

// Protocol constants
const (
	protoVersion = -1
	protoPort    = 40000
)

// handle handles a network connection and chats with the client.
func handle(conn net.Conn) {
	defer conn.Close()
	defer log.Println("client", conn.RemoteAddr(), "disconnected")

	buf := make([]byte, 128)
	for {
		n, err := conn.Read(buf)

		s := string(buf[0:n])
	parse:
		for _, r := range s {
			if unicode.IsSpace(r) {
				continue
			}

			switch r {
			case '\n':
				break parse
			case 'q':
				// We must handle QUIT here to avoid closing the connection both
				// in the deferred call and in cmd.go.
				return
			case 'A', 'a', 'u', 'v':
				// 'Simple' commands not operating on selections.

				args := strings.Fields(s)

				fn := simpleCmdtab[r]
				err := fn(conn, args)
				if err != nil {
					log.Printf("cmd %c: %v", r, err)
				}
				break parse
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Println("conn.Read error:", err)
			}

			return
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
