package chat

type CommandID int

const (
	CmdNick CommandID = iota
	CmdRoom
	CmdRooms
	CmdQuit
	CmdMsg
	CmdUsers
)

type Command struct {
	ID     CommandID
	Client *Client
	Args   []string
}
