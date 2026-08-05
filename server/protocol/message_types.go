package protocol

const (
	// CurrentVersion is the only supported protocol version.
	CurrentVersion = 1
)

const (
	TypeMessage               = "message"
	TypeGroupMessage          = "group_message"
	TypeSenderKeyDistribution = "sender_key_distribution"
	TypeError                 = "error"
)

var supportedTypes = map[string]bool{
	TypeMessage:               true,
	TypeGroupMessage:          true,
	TypeSenderKeyDistribution: true,
}
