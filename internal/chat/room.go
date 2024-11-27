package chat

import "net"

type Room struct {
	Name    string
	Members map[net.Addr]*Client
}

func (r *Room) Broadcast(sender *Client, msg string) {
	for addr, member := range r.Members {
		if addr != sender.Conn.RemoteAddr() {
			member.Msg(msg)
		}
	}
}
