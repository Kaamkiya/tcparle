package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	Conn     net.Conn
	Nick     string
	Room     *Room
	Commands chan<- Command
}

func (c *Client) readInput() {
	reader := bufio.NewReader(c.Conn)

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
			c.Commands <- Command{
				ID:     CmdNick,
				Client: c,
				Args:   args,
			}
		case "/room":
			c.Commands <- Command{
				ID:     CmdRoom,
				Client: c,
				Args:   args,
			}
		case "/rooms":
			c.Commands <- Command{
				ID:     CmdRooms,
				Client: c,
				Args:   args,
			}
		case "/delroom":
			c.Commands <- Command{
				ID:     CmdDelRoom,
				Client: c,
				Args:   args,
			}
		case "/quit":
			c.Commands <- Command{
				ID:     CmdQuit,
				Client: c,
				Args:   args,
			}
		case "/users":
			c.Commands <- Command{
				ID:     CmdUsers,
				Client: c,
				Args:   args,
			}
		default:
			if cmd[0] == '/' {
				c.Err(fmt.Errorf("No such command %s", cmd))
			} else {
				c.Commands <- Command{
					ID:     CmdMsg,
					Client: c,
					Args:   args,
				}
			}
		}
	}
}

func (c *Client) Err(err error) {
	c.Conn.Write([]byte(err.Error() + "\n"))
}

func (c *Client) Msg(msg string) {
	c.Conn.Write([]byte(msg + "\n"))
}
