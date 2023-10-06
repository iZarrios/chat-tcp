package tcpchat

import (
	"bufio"
	"bytes"
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
		rdr := bufio.NewReader(c.Conn)
		msg, err := rdr.ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0]) // cmd should be first token

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
		case "":
			continue
		default:
			c.Cmds <- Cmd{
				ID:     CMD_ERROR,
				Client: c,
				Args:   args,
			}
			// local cmds checker
			var buf bytes.Buffer
			buf.WriteString("INTERNAL: Unknown cmd: ")
			buf.WriteString(cmd)
			buf.WriteString("\n")

			// Get the message as a byte slice
			msgBytes := buf.Bytes()

			c.SendMsgToClient(msgBytes)
		}
	}
}

func (c *Client) SendMsgToClient(msg []byte) error {
	var buf bytes.Buffer
	buf.WriteString("> ")
	buf.Write(msg)
	_, err := c.Conn.Write(buf.Bytes())
	return err
}
