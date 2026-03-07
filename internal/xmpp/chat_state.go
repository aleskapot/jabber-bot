package xmpp

// ChatState represents XEP-0085 chat state notifications
type ChatState string

const (
	ChatStateActive    ChatState = "active"
	ChatStateComposing ChatState = "composing"
	ChatStatePaused    ChatState = "paused"
	ChatStateInactive  ChatState = "inactive"
	ChatStateGone      ChatState = "gone"
)
