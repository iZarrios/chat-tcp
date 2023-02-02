package tcpchat

import (
	"log"
)

func Tcpchat() {
	s := NewServer(":3000")

	// go func() {
	// 	for e := range s.msgChannel {
	// 		fmt.Print(string(e))
	// 	}
	// }()

	err := s.Start()
	if err != nil {
		log.Fatal("Could not start server ", err)
	}
}
