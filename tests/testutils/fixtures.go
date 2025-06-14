package testutils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/models"
)

// TestFixtures provides test data and helper methods
type TestFixtures struct {
	logger *TestLogger
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures(logger *TestLogger) *TestFixtures {
	return &TestFixtures{logger: logger}
}

// CreateTestNode creates a test node with predictable ID
func (f *TestFixtures) CreateTestNode(port int, idSuffix string) *models.Node {
	// Generate a valid hex ID for testing
	nodeID := f.GenerateValidHexID(idSuffix)

	node := &models.Node{
		ID:       nodeID,
		IP:       "127.0.0.1",
		Port:     port,
		LastSeen: time.Now().Unix(),
	}

	f.logger.Info("Created test node: ID=%s..., Port=%d", node.ID[:8], port)
	return node
}

// CreateTestNodes creates multiple test nodes
func (f *TestFixtures) CreateTestNodes(count int, startPort int) []*models.Node {
	nodes := make([]*models.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = f.CreateTestNode(startPort+i, fmt.Sprintf("%03d", i))
	}
	f.logger.Info("Created %d test nodes", count)
	return nodes
}

// CreateTestRoutingTable creates a routing table with test data
func (f *TestFixtures) CreateTestRoutingTable(nodeID string) *models.RoutingTable {
	rt := kademlia.NewRoutingTable(nodeID)
	f.logger.Info("Created test routing table for node %s...", nodeID[:8])
	return rt
}

// CreateTestKeyValueStore creates a KV store with test data
func (f *TestFixtures) CreateTestKeyValueStore(testData map[string]string) *models.KeyValueStore {
	store := kademlia.NewKeyValueStore()

	for key, value := range testData {
		store.Set(key, value)
	}

	f.logger.Info("Created test KV store with %d entries", len(testData))
	return store
}

// GetTestKeyValuePairs returns standard test key-value pairs
func (f *TestFixtures) GetTestKeyValuePairs() map[string]string {
	return map[string]string{
		"1234567890abcdef1234567890abcdef12345678": "test-value-1",
		"abcdef1234567890abcdef1234567890abcdef12": "test-value-2",
		"567890abcdef1234567890abcdef1234567890ab": "test-value-3",
	}
}

// GenerateValidHexID generates a valid 40-character hex ID
func (f *TestFixtures) GenerateValidHexID(suffix string) string {
	// Generate a valid hex ID similar to the actual implementation
	data := fmt.Sprintf("test-%s-%d", suffix, time.Now().UnixNano())
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateInvalidIDs returns various invalid ID formats for testing
func (f *TestFixtures) GenerateInvalidIDs() map[string]string {
	return map[string]string{
		"too_short":    "123abc",
		"too_long":     "1234567890abcdef1234567890abcdef123456789extra",
		"non_hex":      "1234567890abcdef1234567890abcdef1234567g",
		"empty":        "",
		"spaces":       "1234567890abcdef 234567890abcdef12345678",
		"with_symbols": "1234567890abcdef1234567890abcdef1234567!",
	}
}

// CreatePopulatedRoutingTable creates a routing table with several test nodes
func (f *TestFixtures) CreatePopulatedRoutingTable(localNodeID string, nodeCount int) *models.RoutingTable {
	rt := kademlia.NewRoutingTable(localNodeID)

	// Add test nodes
	testNodes := f.CreateTestNodes(nodeCount, 8080)
	for _, node := range testNodes {
		kademlia.AddNodeToRoutingTable(rt, node, localNodeID)
	}

	f.logger.Info("Created populated routing table with %d nodes", nodeCount)
	return rt
}
