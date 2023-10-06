package tcpchat

import (
	"log"
)

func Tcpchat() {
	s := NewServer(":3000")

	err := s.Start()
	if err != nil {
		log.Fatal("Could not start server ", err)
	}
}
