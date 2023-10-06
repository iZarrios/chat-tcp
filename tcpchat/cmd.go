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
const (
	RED   = "\033[31m"
	RESET = "\033[0m"
)

type CmdID int

type Cmd struct {
	ID     CmdID
	Client *Client
	Args   []string
}
