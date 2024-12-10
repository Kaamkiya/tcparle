package chat

import "net"

type room struct {
	Name    string
	Members map[net.Addr]*client
}

func (r *room) Broadcast(sender *client, msg string) {
	for addr, member := range r.Members {
		if addr != sender.conn.RemoteAddr() {
			member.msg(msg)
		}
	}
}
