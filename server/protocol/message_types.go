package protocol

const (
	// CurrentVersion is the only supported protocol version.
	CurrentVersion = 1
)

const (
	TypeMessage               = "message"
	TypeGroupMessage          = "group_message"
	TypeSenderKeyDistribution = "sender_key_distribution"
	TypeProfilePictureUpdated = "profile_picture_updated"
	// TypePresence and TypePresenceSnapshot are server→client only. They are
	// NOT added to supportedTypes so incoming envelopes with these types are
	// rejected as invalid_message_type.
	TypePresence         = "presence"
	TypePresenceSnapshot = "presence_snapshot"
	TypeError            = "error"
)

var supportedTypes = map[string]bool{
	TypeMessage:               true,
	TypeGroupMessage:          true,
	TypeSenderKeyDistribution: true,
	TypeProfilePictureUpdated: true,
}
