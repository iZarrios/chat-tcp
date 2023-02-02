package tcpchat

import (
	"fmt"
	"net"
)

type Server struct {
	listenAddress string
	ln            net.Listener
	Rooms         map[string]*Room
	Cmds          chan Cmd
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		Rooms:         make(map[string]*Room),
		Cmds:          make(chan Cmd),
	}

}
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

}

func (s *Server) AcceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Could not accept: ", err)
			continue
		}
		// read here(init client)
		fmt.Printf("%v Connected\n", conn.RemoteAddr())
	}
}
