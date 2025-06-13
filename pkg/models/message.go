package models

type MessageType string

const (
	Ping      MessageType = "PING"
	FindNode  MessageType = "FIND_NODE"
	Store     MessageType = "STORE"
	FindValue MessageType = "FIND_VALUE"
	Pong      MessageType = "PONG"
)

type Message struct {
	Type   MessageType // Type of the message (PING, STORE, etc.)
	Sender Node        // Sender's information
	Key    string      // Key being looked up (if applicable)
	Value  string      // Value to store (if applicable)
	Target string      // Target ID for FIND_NODE or FIND_VALUE
}
