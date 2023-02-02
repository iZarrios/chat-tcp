package tcpchat

type CmdID int
type Cmd struct {
	ID     CmdID
	Client *Client
	Args   []string
}
