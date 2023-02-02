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
	quitChannel   chan struct{}
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		Rooms:         make(map[string]*Room),
		Cmds:          make(chan Cmd), //TODO: make it buffered?
		quitChannel:   make(chan struct{}),
	}

}
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.AcceptLoop()

	go s.RunCmd()

	<-s.quitChannel

	return nil

}

func (s *Server) AcceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Could not accept: ", err)
			continue
		}
		// read or init client
		c := s.NewClient(conn)
		go c.ReadLoop()

		fmt.Printf("%v Connected\n", conn.RemoteAddr())

	}
}

func (s *Server) NewClient(conn net.Conn) *Client {
	fmt.Printf("%v has Connected Successfully\n", conn.RemoteAddr().String())
	c := &Client{
		Conn: conn,
		Nick: "anon",
		Room: nil,    // user is not in any Room at the beginning
		Cmds: s.Cmds, // read only commands
	}

	return c
}

func (s *Server) RunCmd() {
	for c := range s.Cmds {
		switch c.ID {
		case CMD_JOIN:
			fmt.Println("got join")
		case CMD_NICK:
			fmt.Println("got nick")
		case CMD_ROOMS:
			fmt.Println("got rooms")
		case CMD_MSG:
			fmt.Println("got msg")
		case CMD_QUIT:
			fmt.Println("got quit")
		default:
			continue

		}
	}
}
