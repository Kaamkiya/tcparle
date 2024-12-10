package chat

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
)

type server struct {
	rooms    map[string]*room
	commands chan command
}

func NewServer() *server {
	srv := &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
	}

	srv.rooms["main"] = &room{
		Name:    "main",
		Members: make(map[net.Addr]*client),
	}

	return srv
}

func (srv *server) NewClient(conn net.Conn) {
	log.Printf("New connection from %s", conn.RemoteAddr())

	c := &client{
		conn:     conn,
		nick:     "Guest" + strconv.Itoa(rand.IntN(100000)),
		commands: srv.commands,
	}

	c.commands <- command{
		id:     CmdRoom,
		client: c,
		args:   []string{"/room", "main"},
	}

	c.readInput()
}

func (srv *server) RunCommands() {
	for cmd := range srv.commands {
		switch cmd.id {
		case CmdNick:
			srv.cmdNick(cmd.args, cmd.client)
		case CmdQuit:
			srv.cmdQuit(cmd.args, cmd.client)
		case CmdRoom:
			srv.cmdRoom(cmd.args, cmd.client)
		case CmdRooms:
			srv.cmdRooms(cmd.args, cmd.client)
		case CmdDelRoom:
			srv.cmdDelroom(cmd.args, cmd.client)
		case CmdMsg:
			srv.cmdMsg(cmd.args, cmd.client)
		case CmdUsers:
			srv.cmdUsers(cmd.args, cmd.client)
		}
	}
}

func (srv *server) cmdNick(args []string, c *client) {
	if len(args) == 2 {
		newNick := args[1]
		for _, char := range newNick {
			if !strings.ContainsRune(
				"qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890",
				char,
			) {
				c.err(fmt.Errorf("Nickname must consist of only numbers and letters."))
				return
			}
		}
		if c.room != nil {
			c.room.Broadcast(c, fmt.Sprintf("%s is now %s.", c.nick, newNick))
		}
		c.nick = newNick
		c.msg("You have been renamed.")
	} else {
		c.err(fmt.Errorf("Usage: /nick [name]"))
	}
}

func (srv *server) cmdRooms(_ []string, c *client) {
	c.msg("Here's a list of all the rooms:")
	for roomName := range srv.rooms {
		c.msg(roomName)
	}
}

func (srv *server) cmdRoom(args []string, c *client) {
	if len(args) < 2 {
		c.msg(fmt.Sprintf("You're currently in %s", c.room.Name))
		return
	}
	roomName := args[1]
	r, ok := srv.rooms[roomName]

	if !ok {
		r = &room{
			Name:    roomName,
			Members: make(map[net.Addr]*client),
		}

		srv.rooms[roomName] = r
	}

	r.Members[c.conn.RemoteAddr()] = c
	srv.quitCurrentroom(c)
	c.room = r
	r.Broadcast(c, fmt.Sprintf("%s has joined the room.", c.nick))
	c.msg(
		fmt.Sprintf(
			"You are now in %s. There are %d other users.",
			r.Name,
			len(r.Members)-1, // One person in this room is the current client. Don't count them.
		),
	)
}

func (srv *server) cmdDelroom(args []string, c *client) {
	if len(args) < 2 {
		c.err(fmt.Errorf("Usage: /delroom [room]"))
		return
	}

	r, ok := srv.rooms[args[1]]

	if !ok {
		c.err(fmt.Errorf("No such room."))
		return
	}

	if r.Name == "main" {
		c.err(fmt.Errorf("Cannot delete main."))
	}

	if len(r.Members) == 0 {
		delete(srv.rooms, r.Name)
	} else {
		c.err(fmt.Errorf("There are still people in that room."))
	}
}

func (srv *server) cmdQuit(args []string, c *client) {
	c.msg("Goodbye.")
	srv.quitCurrentroom(c)
	log.Printf("Client %s has disconnected.", c.conn.RemoteAddr())
	c.conn.Close()
}

func (srv *server) cmdMsg(args []string, c *client) {
	if c.room == nil {
		c.err(fmt.Errorf("You must be in a room."))
		return
	}

	c.room.Broadcast(c, c.nick+": "+strings.Join(args, " "))
}

func (srv *server) cmdUsers(args []string, c *client) {
	if c.room == nil {
		c.err(fmt.Errorf("You must be in a room."))
		return
	}

	c.msg("Users: ")
	for _, user := range c.room.Members {
		c.msg(fmt.Sprintf("* %s", user.nick))
	}
}

func (srv *server) quitCurrentroom(c *client) {
	if c.room != nil {
		delete(c.room.Members, c.conn.RemoteAddr())
		c.room.Broadcast(c, fmt.Sprintf("%s has left the room.", c.nick))
	}
}
