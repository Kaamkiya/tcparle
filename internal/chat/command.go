package chat

type commandID int

const (
	CmdNick commandID = iota
	CmdRoom
	CmdRooms
	CmdDelRoom
	CmdQuit
	CmdMsg
	CmdUsers
)

type command struct {
	id     commandID
	client *client
	args   []string
}
