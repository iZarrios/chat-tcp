package tcpchat

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Server struct {
	ListenAddress string
	ln            net.Listener
	Rooms         map[string]*Room
	Cmds          chan Cmd
	quitChannel   chan struct{}
}

type ServerOption func(*Server)

func WithAddress(listenAddress string) ServerOption {
	return func(s *Server) {
		s.ListenAddress = listenAddress
	}
}

func WithCmdBufferSize(bufferSize int) ServerOption {
	return func(s *Server) {
		if s.Cmds != nil {
			close(s.Cmds)
		}

		s.Cmds = make(chan Cmd, bufferSize)
	}
}

func NewServer(opts ...ServerOption) *Server {
	const (
		defaultBufferSize    = 1
		defaultListenAddress = ":3000"
		defaultRoomWatcher   = ""
	)
	server := &Server{
		ListenAddress: defaultListenAddress,
		Rooms:         make(map[string]*Room),
		Cmds:          make(chan Cmd, defaultBufferSize),
		quitChannel:   make(chan struct{}),
	}

	// Loop through each option
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.AcceptLoop()

	go s.RunCmd()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c // to make it unblocking
		s.quitChannel <- struct{}{}
		os.Exit(1)
	}()

	<-s.quitChannel

	return nil

}

func (s *Server) AcceptLoop() {
	for {
		conn, err := s.ln.Accept()

		if err != nil {
			fmt.Printf("Could not accept %v: %v \n",
				conn.RemoteAddr().String(), err)
			continue
		}
		// read or init client
		c := s.NewClient(conn)
		go c.ReadLoop()
		fmt.Printf("%v Connected\n", conn.RemoteAddr())

	}
}

func (s *Server) NewClient(conn net.Conn) *Client {
	// fmt.Printf("%v has Connected Successfully\n", conn.RemoteAddr().String())
	c := &Client{
		Conn: conn,
		Nick: "anon",
		Room: nil,    // user is not in any Room at the beginning
		Cmds: s.Cmds, // read only commands
	}

	return c
}

func (s *Server) RunCmd() {
	for cmd := range s.Cmds {
		switch cmd.ID {
		case CMD_JOIN:
			s.join(cmd.Client, cmd.Args)
		case CMD_NICK:
			s.changeNickname(cmd.Client, cmd.Args)
		case CMD_ROOMS:
			if len(s.Rooms) == 0 {
				s.SendMsgToClient(cmd.Client, []byte("Look like there are no rooms available u can /join NEW_ROOM_NAME and it will get created!\n"))
			}
			for k, r := range s.Rooms {
				var buf bytes.Buffer
				_, err := fmt.Fprintf(&buf, "%4v|%4v|%4vMembers\n", k, r.Name, len(r.Members))

				if err != nil {
					panic("Writing into buf when trying to RunCmd failed")
				}

				s.SendMsgToClient(cmd.Client, buf.Bytes())
			}
		case CMD_MSG:
			if cmd.Client.Room != nil {
				s.SendMsgToRoom(cmd.Client, cmd.Args)
			} else {
				s.SendMsgToClient(cmd.Client, []byte("You need to join a room first try /rooms\n"))
			}
		case CMD_QUIT:
			s.QuitChatApp(cmd.Client)
		case CMD_ERROR:
			fmt.Print(RED)
			fmt.Printf("[ERROR] Cmd Error format from %v(%v)\n", cmd.Client.Nick, cmd.Client.Conn.RemoteAddr().String())
			fmt.Print(RESET)
			s.BadCmd(cmd.Client, cmd.Args[0])
		default:
			fmt.Print("default in server?\n")

		}
	}
}

func (s *Server) join(c *Client, args []string) {
	if len(args) != 2 {
		c.SendMsgToClient([]byte("room name is required. usage: /join ROOM_NAME"))
		return
	}

	roomName := args[1]

	r, ok := s.Rooms[roomName]
	if !ok {
		r = &Room{
			Name:    roomName,
			Members: make(map[net.Addr]*Client),
		}
		s.Rooms[roomName] = r
	}
	r.Members[c.Conn.RemoteAddr()] = c

	s.QuitCurrentRoom(c)
	c.Room = r

	r.Broadcast(c, []byte(fmt.Sprintf("%s joined the room", c.Nick)))

	c.SendMsgToClient([]byte(fmt.Sprintf("welcome to %s\n", roomName)))
}

func (s *Server) changeNickname(c *Client, args []string) {
	if len(args) == 2 {
		c.Room.Broadcast(c, []byte(fmt.Sprintf("%v has been changed his nicknames into %v", c.Nick, args[1])))

		s.SendMsgToClient(c, []byte(fmt.Sprintf("Successfully changed Nick into %v\n", args[1])))
		c.Nick = args[1]
		return
	}
	s.BadCmd(c, args[0])

}

func (s *Server) SendMsgToRoom(c *Client, args []string) {
	if len(args) < 2 {
		c.SendMsgToClient([]byte("message is required, usage: /msg MSG\n"))
		return
	}

	var buf bytes.Buffer
	buf.WriteString(c.Nick)
	buf.WriteString(": ")
	buf.WriteString(strings.Join(args[1:], " "))
	c.Room.Broadcast(c, buf.Bytes())
}

func (s *Server) BadCmd(c *Client, cmdName string) {
	var buf bytes.Buffer
	buf.WriteString("Unknown cmd: ")
	buf.WriteString(cmdName)
	buf.WriteRune('\n')
	s.SendMsgToClient(c, buf.Bytes())

	log.Printf(buf.String())

}

func (s *Server) SendMsgToClient(c *Client, msg []byte) {
	_, err := c.Conn.Write([]byte(msg))

	if err != nil {
		fmt.Printf("SendMsgToClient Error")
	}
}

func (s *Server) QuitCurrentRoom(c *Client) {
	if c.Room != nil {
		oldRoom := s.Rooms[c.Room.Name]
		delete(s.Rooms[c.Room.Name].Members, c.Conn.RemoteAddr())
		var buf bytes.Buffer
		buf.WriteString(c.Nick)
		buf.WriteString(" has left the room")
		oldRoom.Broadcast(c, buf.Bytes())
		// deleting room if there are no members in the room
		if len(s.Rooms[c.Room.Name].Members) == 0 {
			delete(s.Rooms, oldRoom.Name)
		}
	}
}

func (s *Server) QuitChatApp(c *Client) {
	fmt.Printf("client has left the chat: %s\n", c.Conn.RemoteAddr().String())

	s.QuitCurrentRoom(c)

	c.SendMsgToClient([]byte("See you soon!\n"))
	c.Conn.Close()
}
