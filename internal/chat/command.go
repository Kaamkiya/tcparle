package chat

type CommandID int

const (
	CmdNick CommandID = iota
	CmdRoom
	CmdQuit
	CmdMsg
	CmdUsers
)

type Command struct {
	ID CommandID
	Client *Client
	Args []string
}
