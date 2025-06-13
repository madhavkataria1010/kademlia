package models

type Node struct {
	ID       string // Unique identifier for the node (e.g., SHA-1 or XOR hash of IP+port)
	IP       string // IP address of the node
	Port     int    // Port on which the node is listening
	LastSeen int64  // Timestamp for when the node was last active
}

type Bucket struct {
	Nodes   []*Node // List of nodes in the bucket
	MaxSize int     // Maximum allowed nodes (k)
}

type RoutingTable struct {
	Buckets []*Bucket // List of buckets
}
