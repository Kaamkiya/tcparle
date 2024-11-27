package chat

type CommandID int

const (
	CmdNick CommandID = iota
	CmdRoom
	CmdQuit
	CmdMsg
)

type Command struct {
	ID CommandID
	Client *Client
	Args []string
}
