package tcp

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	listenAddress string
	ln            net.Listener
	quitChannel   chan struct{}
	msgChannel    chan []byte
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		quitChannel:   make(chan struct{}),
		msgChannel:    make(chan []byte, 10),
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

	// wait until quitChannel is not blocked (always blocked since un-buffered channel)
	<-s.quitChannel
	close(s.quitChannel)
	close(s.msgChannel)

	return nil
}

func (s *Server) AcceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Could not accept: ", err)
			continue
		}
		fmt.Printf("%v Connected\n", conn.RemoteAddr())

		go s.ReadLoop(conn)
	}
}

func (s *Server) ReadLoop(conn net.Conn) error {
	defer conn.Close()

	buf := make([]byte, 2048)
	buffName := make([]byte, 10)

	conn.Write([]byte("What do you want to be called? "))
	n, err := conn.Read(buffName)
	if err != nil {
		return err
	}
	name := string(buffName[:n])
	fmt.Printf("%vhas joined the chat!\n", name)
	for {
		n, err := conn.Read(buf)

		if err != nil {
			return err
		}
		s.msgChannel <- buf[:n]
	}
}

func Tcp() {
	s := NewServer(":3000")

	go func() {
		for e := range s.msgChannel {
			fmt.Print(string(e))
		}
	}()

	err := s.Start()
	if err != nil {
		log.Fatal("Could not start server ", err)
	}

}
