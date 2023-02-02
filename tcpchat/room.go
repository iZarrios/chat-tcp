package tcpchat

import "net"

type Room struct {
	Name    string
	Members map[net.Addr]*Client
}
