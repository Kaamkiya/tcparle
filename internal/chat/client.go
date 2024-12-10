package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	nick     string
	room     *room
	commands chan<- command
}

func (c *client) readInput() {
	reader := bufio.NewReader(c.conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			// The likely only way you won't get a \n is an EOF, in
			// which case we exit.
			return
		}

		msg = strings.Trim(msg, "\r\n")

		if len(msg) == 0 {
			// We don't need to do anything if the client entered
			// and empty line. Also, if we tried to split it, we
			// would get an error, so we skip it entirely.
			continue
		}

		// The command name shouldn't be counted as an argument.
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/nick":
			c.commands <- command{
				id:     CmdNick,
				client: c,
				args:   args,
			}
		case "/room":
			c.commands <- command{
				id:     CmdRoom,
				client: c,
				args:   args,
			}
		case "/rooms":
			c.commands <- command{
				id:     CmdRooms,
				client: c,
				args:   args,
			}
		case "/delroom":
			c.commands <- command{
				id:     CmdDelRoom,
				client: c,
				args:   args,
			}
		case "/quit":
			c.commands <- command{
				id:     CmdQuit,
				client: c,
				args:   args,
			}
		case "/users":
			c.commands <- command{
				id:     CmdUsers,
				client: c,
				args:   args,
			}
		default:
			if cmd[0] == '/' {
				c.err(fmt.Errorf("No such command %s", cmd))
			} else {
				c.commands <- command{
					id:     CmdMsg,
					client: c,
					args:   args,
				}
			}
		}
	}
}

func (c *client) err(err error) {
	c.conn.Write([]byte(err.Error() + "\n"))
}

func (c *client) msg(msg string) {
	c.conn.Write([]byte(msg + "\n"))
}
