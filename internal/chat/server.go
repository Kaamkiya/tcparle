package chat

import (
	"log"
	"strconv"
	"strings"
	"net"
	"fmt"
	"math/rand"
)

type Server struct {
	Rooms map[string]*Room
	Commands chan Command
}

func NewServer() *Server {
	srv := &Server {
		Rooms: make(map[string]*Room),
		Commands: make(chan Command),
	}

	srv.Rooms["main"] = &Room{
		Name: "main",
		Members: make(map[net.Addr]*Client),
	}

	return srv
}

func (srv *Server) NewClient(conn net.Conn) {
	log.Printf("New connection from %s", conn.RemoteAddr())

	c := &Client{
		Conn: conn,
		Nick: "Guest"+strconv.Itoa(rand.Intn(10000)),
		Commands: srv.Commands,
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
		case CmdMsg:
			srv.cmdMsg(cmd.Args, cmd.Client)
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
		c.Room.Broadcast(c, fmt.Sprintf("%s is now %s.", c.Nick, newNick))
		c.Nick = newNick
		c.Msg("You have been renamed.")
	} else {
		c.Err(fmt.Errorf("Usage: /nick [name]"))
	}
}

func (srv *Server) cmdRoom(args []string, c *Client) {
	if len(args) < 2 {
		c.Msg("Here's a list of all the rooms:")
		for roomName, _ := range srv.Rooms {
			c.Msg(roomName)
		}
		return
	}

	roomName := args[1]
	r, ok := srv.Rooms[roomName]

	if !ok {
		r = &Room{
			Name: roomName,
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
			len(r.Members) - 1, // One person in this room is the current client. Don't count them.
		),
	)
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

func (srv *Server) quitCurrentRoom(c *Client) {
	if c.Room != nil {
		delete(c.Room.Members, c.Conn.RemoteAddr())
		c.Room.Broadcast(c, fmt.Sprintf("%s has left the room.", c.Nick))
	}
}
