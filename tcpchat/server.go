package tcpchat

import (
	"fmt"
	"log"
	"net"
	"strings"
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
	for cmd := range s.Cmds {
		switch cmd.ID {
		case CMD_JOIN:
			s.join(cmd.Client, cmd.Args)
		case CMD_NICK:
			s.changeNickname(cmd.Client, cmd.Args)
		case CMD_ROOMS:
			if len(s.Rooms) == 0 {
				s.SendMsgToClient(*cmd.Client, "Look like there are no rooms available u can /join NEW_ROOM_NAME and it will get created!\n")
			}
			for k, r := range s.Rooms {
				s.SendMsgToClient(*cmd.Client, fmt.Sprintf("%v | %v | %v Members\n",
					k, r.Name, len(r.Members)))
			}
		case CMD_MSG:
			if cmd.Client.Room != nil {
				s.SendMsgToRoom(cmd.Client, cmd.Args)
			} else {
				s.SendMsgToClient(*cmd.Client, "You need to join a room first try /rooms\n")
			}
		case CMD_QUIT:
			s.QuitChatApp(cmd.Client)
		case CMD_ERROR:
			fmt.Println("Cmd Error format")
			s.BadCmd(*cmd.Client, cmd.Args[0])
		default:
			fmt.Print("default in server?\n")

		}
	}
}
func (s *Server) join(c *Client, args []string) {
	if len(args) != 2 {
		c.SendMsgToClient("room name is required. usage: /join ROOM_NAME")
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

	r.Broadcast(c, fmt.Sprintf("%s joined the room", c.Nick))

	c.SendMsgToClient(fmt.Sprintf("welcome to %s\n", roomName))
}

func (s *Server) changeNickname(c *Client, args []string) {
	if len(args) == 2 {
		c.Room.Broadcast(c, fmt.Sprintf("%v has been changed his nicknames into %v", c.Nick, args[1]))

		s.SendMsgToClient(*c, fmt.Sprintf("Successfully changed Nick into %v\n", args[1]))
		c.Nick = args[1]
		return
	}
	s.BadCmd(*c, args[0])

}
func (s *Server) SendMsgToRoom(c *Client, args []string) {
	if len(args) < 2 {
		c.SendMsgToClient("message is required, usage: /msg MSG\n")
		return
	}

	msg := strings.Join(args[1:], " ")
	c.Room.Broadcast(c, c.Nick+": "+msg)
}

func (s *Server) BadCmd(c Client, cmdName string) {
	s.SendMsgToClient(c, fmt.Sprintf("Unknown cmd: %v\n", cmdName))

}

func (s *Server) SendMsgToClient(c Client, msg string) {
	_, err := c.Conn.Write([]byte(msg))

	if err != nil {
		fmt.Printf("SendMsgToClient Error")
	}
}

func (s *Server) QuitCurrentRoom(c *Client) {
	if c.Room != nil {
		oldRoom := s.Rooms[c.Room.Name]
		delete(s.Rooms[c.Room.Name].Members, c.Conn.RemoteAddr())
		oldRoom.Broadcast(c, fmt.Sprintf("%s has left the room", c.Nick))
	}
}

func (s *Server) QuitChatApp(c *Client) {
	log.Printf("client has left the chat: %s\n", c.Conn.RemoteAddr().String())

	s.QuitCurrentRoom(c)

	c.SendMsgToClient("See you soon!\n")
	c.Conn.Close()
}
