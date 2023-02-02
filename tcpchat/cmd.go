package tcpchat

//https://www.tutorialspoint.com/how-to-use-iota-in-golang

const (
	CMD_NICK CmdID = iota
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
	CMD_ERROR
)

type CmdID int

type Cmd struct {
	ID     CmdID
	Client *Client
	Args   []string
}
