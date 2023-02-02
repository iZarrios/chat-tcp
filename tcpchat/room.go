package tcpchat

import "net"

type Room struct {
	Name    string
	Members map[net.Addr]*Client
}

func (r *Room) Broadcast(sender *Client, msg string) {
	for addr, m := range r.Members {
		if sender.Conn.RemoteAddr() != addr {
			m.SendMsgToClient(msg + "\n")
		}
	}
}
