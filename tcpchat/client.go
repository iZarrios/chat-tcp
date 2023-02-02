package tcpchat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	Conn net.Conn
	Nick string
	Room *Room
	Cmds chan<- Cmd
}

// equivalent of ReadLoop() in simple tcp
// read what the user input is then, send it to the server
func (c *Client) ReadLoop() {
	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0]) // Cmd should be first token

		switch cmd {
		case "/join":
			c.Cmds <- Cmd{
				ID:     CMD_JOIN,
				Client: c,
				Args:   args,
			}
		case "/nick":
			c.Cmds <- Cmd{
				ID:     CMD_NICK,
				Client: c,
				Args:   args,
			}
		case "/rooms":
			c.Cmds <- Cmd{
				ID:     CMD_ROOMS,
				Client: c,
				Args:   args,
			}
		case "/msg":
			c.Cmds <- Cmd{
				ID:     CMD_MSG,
				Client: c,
				Args:   args,
			}
		case "/quit":
			c.Cmds <- Cmd{
				ID:     CMD_QUIT,
				Client: c,
				Args:   args,
			}
		default:
			c.Cmds <- Cmd{
				ID:     CMD_ERROR,
				Client: c,
				Args:   args,
			}
			// local cmds checker
			c.SendMsgToClient(fmt.Sprintf("INTERNAL:Unknown cmd: %v", cmd))
			continue

		}
	}

}

func (c *Client) SendMsgToClient(msg string) error {
	_, err := c.Conn.Write([]byte("> " + msg))
	return err
}
