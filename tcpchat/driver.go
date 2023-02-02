package tcpchat

import (
	"log"
	"net"
)

func Tcpchat() {
	s := NewServer()

	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal("unable to start server ", err)
	}
	defer ln.Close()
}
