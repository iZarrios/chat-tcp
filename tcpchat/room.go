package tcpchat

import "net"

type Room struct {
	Name    string
	Members map[net.Addr]*Client
}

func (r *Room) Broadcast(sender *Client, msg []byte) {
	for addr, subscriber := range r.Members {
		if sender.Conn.RemoteAddr() != addr {
			msg = append(msg, '\n')
			subscriber.SendMsgToClient(msg)
		}
	}
}
