package tcpchat

import "net"

type Client struct {
	Conn net.Conn
	Nick string
	Room *Room
	Cmds chan<- Cmd
}

func (c *Client) hi() {

}
