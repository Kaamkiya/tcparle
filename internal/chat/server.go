package chat

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	Rooms    map[string]*Room
	Commands chan Command
}

func NewServer() *Server {
	srv := &Server{
		Rooms:    make(map[string]*Room),
		Commands: make(chan Command),
	}

	srv.Rooms["main"] = &Room{
		Name:    "main",
		Members: make(map[net.Addr]*Client),
	}

	return srv
}

func (srv *Server) NewClient(conn net.Conn) {
	log.Printf("New connection from %s", conn.RemoteAddr())

	c := &Client{
		Conn:     conn,
		Nick:     "Guest" + strconv.Itoa(rand.IntN(100000)),
		Commands: srv.Commands,
	}

	c.Commands <- Command{
		ID:     CmdRoom,
		Client: c,
		Args:   []string{"/room", "main"},
	}

	c.readInput()
}

func (srv *Server) RunCommands() {
	for cmd := range srv.Commands {
		switch cmd.ID {
		case CmdNick:
			srv.cmdNick(cmd.Args, cmd.Client)
		case CmdQuit:
			srv.cmdQuit(cmd.Args, cmd.Client)
		case CmdRoom:
			srv.cmdRoom(cmd.Args, cmd.Client)
		case CmdRooms:
			srv.cmdRooms(cmd.Args, cmd.Client)
		case CmdDelRoom:
			srv.cmdDelRoom(cmd.Args, cmd.Client)
		case CmdMsg:
			srv.cmdMsg(cmd.Args, cmd.Client)
		case CmdUsers:
			srv.cmdUsers(cmd.Args, cmd.Client)
		}
	}
}

func (srv *Server) cmdNick(args []string, c *Client) {
	if len(args) == 2 {
		newNick := args[1]
		for _, char := range newNick {
			if !strings.ContainsRune(
				"qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890",
				char,
			) {
				c.Err(fmt.Errorf("Nickname must consist of only numbers and letters."))
				return
			}
		}
		if c.Room != nil {
			c.Room.Broadcast(c, fmt.Sprintf("%s is now %s.", c.Nick, newNick))
		}
		c.Nick = newNick
		c.Msg("You have been renamed.")
	} else {
		c.Err(fmt.Errorf("Usage: /nick [name]"))
	}
}

func (srv *Server) cmdRooms(_ []string, c *Client) {
	c.Msg("Here's a list of all the rooms:")
	for roomName := range srv.Rooms {
		c.Msg(roomName)
	}
}

func (srv *Server) cmdRoom(args []string, c *Client) {
	if len(args) < 2 {
		c.Msg(fmt.Sprintf("You're currently in %s", c.Room.Name))
		return
	}
	roomName := args[1]
	r, ok := srv.Rooms[roomName]

	if !ok {
		r = &Room{
			Name:    roomName,
			Members: make(map[net.Addr]*Client),
		}

		srv.Rooms[roomName] = r
	}

	r.Members[c.Conn.RemoteAddr()] = c
	srv.quitCurrentRoom(c)
	c.Room = r
	r.Broadcast(c, fmt.Sprintf("%s has joined the room.", c.Nick))
	c.Msg(
		fmt.Sprintf(
			"You are now in %s. There are %d other users.",
			r.Name,
			len(r.Members)-1, // One person in this room is the current client. Don't count them.
		),
	)
}

func (srv *Server) cmdDelRoom(args []string, c *Client) {
	if len(args) < 2 {
		c.Err(fmt.Errorf("Usage: /delroom [room]"))
		return
	}

	r, ok := srv.Rooms[args[1]]

	if !ok {
		c.Err(fmt.Errorf("No such room."))
		return
	}

	if r.Name == "main" {
		c.Err(fmt.Errorf("Cannot delete main."))
	}

	if len(r.Members) == 0 {
		delete(srv.Rooms, r.Name)
	} else {
		c.Err(fmt.Errorf("There are still people in that room."))
	}
	return
}

func (srv *Server) cmdQuit(args []string, c *Client) {
	c.Msg("Goodbye.")
	srv.quitCurrentRoom(c)
	log.Printf("Client %s has disconnected.", c.Conn.RemoteAddr())
	c.Conn.Close()
}

func (srv *Server) cmdMsg(args []string, c *Client) {
	if c.Room == nil {
		c.Err(fmt.Errorf("You must be in a room."))
		return
	}

	c.Room.Broadcast(c, c.Nick+": "+strings.Join(args, " "))
}

func (srv *Server) cmdUsers(args []string, c *Client) {
	if c.Room == nil {
		c.Err(fmt.Errorf("You must be in a room."))
		return
	}

	c.Msg("Users: ")
	for _, user := range c.Room.Members {
		c.Msg(fmt.Sprintf("* %s", user.Nick))
	}
}

func (srv *Server) quitCurrentRoom(c *Client) {
	if c.Room != nil {
		delete(c.Room.Members, c.Conn.RemoteAddr())
		c.Room.Broadcast(c, fmt.Sprintf("%s has left the room.", c.Nick))
	}
}
