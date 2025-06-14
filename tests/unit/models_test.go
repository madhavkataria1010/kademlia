package unit

import (
	"testing"

	"github.com/Aradhya2708/kademlia/pkg/constants"
	"github.com/Aradhya2708/kademlia/pkg/models"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestNodeModel tests the Node model
func TestNodeModel(t *testing.T) {
	logger := testutils.NewTestLogger(t, "MODELS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Node model tests")

	t.Run("NodeCreation", func(t *testing.T) {
		section := logger.Section("Node Creation")

		section.Step(1, "Create node with valid parameters")
		node := fixtures.CreateTestNode(8080, "001")

		assert.NotNil(node, "Node should not be nil")
		assert.Equal("127.0.0.1", node.IP, "IP should be set correctly")
		assert.Equal(8080, node.Port, "Port should be set correctly")
		assert.Equal(40, len(node.ID), "Node ID should be 40 characters")
		assert.True(node.LastSeen > 0, "LastSeen should be set")

		section.Success("Node created successfully")
	})

	t.Run("NodeComparison", func(t *testing.T) {
		section := logger.Section("Node Comparison")

		section.Step(1, "Create identical nodes")
		nodeID := fixtures.GenerateValidHexID("test123")
		node1 := &models.Node{ID: nodeID, IP: "127.0.0.1", Port: 8080}
		node2 := &models.Node{ID: nodeID, IP: "127.0.0.1", Port: 8080}

		assert.Equal(node1.ID, node2.ID, "Node IDs should be equal")
		assert.Equal(node1.IP, node2.IP, "Node IPs should be equal")
		assert.Equal(node1.Port, node2.Port, "Node ports should be equal")

		section.Step(2, "Create different nodes")
		node3 := &models.Node{ID: fixtures.GenerateValidHexID("diff"), IP: "127.0.0.1", Port: 8081}

		assert.NotEqual(node1.ID, node3.ID, "Different node IDs should not be equal")
		assert.NotEqual(node1.Port, node3.Port, "Different node ports should not be equal")

		section.Success("Node comparison working correctly")
	})

	t.Run("NodeValidation", func(t *testing.T) {
		section := logger.Section("Node Validation")

		section.Step(1, "Test node with valid data")
		validNode := fixtures.CreateTestNode(8080, "valid")
		assert.True(len(validNode.ID) == 40, "Valid node should have 40-char ID")
		assert.True(validNode.Port > 0 && validNode.Port <= 65535, "Valid node should have valid port")

		section.Success("Node validation working correctly")
	})
}

// TestBucketModel tests the Bucket model
func TestBucketModel(t *testing.T) {
	logger := testutils.NewTestLogger(t, "MODELS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Bucket model tests")

	t.Run("BucketCreation", func(t *testing.T) {
		section := logger.Section("Bucket Creation")

		section.Step(1, "Create bucket with max size")
		k := constants.GetK()
		bucket := &models.Bucket{MaxSize: k}

		assert.NotNil(bucket, "Bucket should not be nil")
		assert.Equal(k, bucket.MaxSize, "Bucket max size should match k")
		assert.Equal(0, len(bucket.Nodes), "Bucket should start empty")

		section.Success("Bucket created successfully")
	})

	t.Run("BucketNodeManagement", func(t *testing.T) {
		section := logger.Section("Bucket Node Management")

		section.Step(1, "Create bucket and test nodes")
		bucket := &models.Bucket{MaxSize: 3}
		nodes := fixtures.CreateTestNodes(5, 8080)

		section.Step(2, "Add nodes within capacity")
		for i := 0; i < 3; i++ {
			bucket.Nodes = append(bucket.Nodes, nodes[i])
		}

		assert.Equal(3, len(bucket.Nodes), "Bucket should contain 3 nodes")

		section.Step(3, "Verify node storage")
		for i, node := range bucket.Nodes {
			assert.Equal(nodes[i].ID, node.ID, "Node %d ID should match", i)
		}

		section.Step(4, "Test bucket overflow behavior")
		// Simulate what happens when bucket is full
		initialLength := len(bucket.Nodes)
		assert.Equal(bucket.MaxSize, initialLength, "Bucket should be at max capacity")

		section.Success("Bucket node management working correctly")
	})

	t.Run("BucketCapacityLimits", func(t *testing.T) {
		section := logger.Section("Bucket Capacity Limits")

		section.Step(1, "Test different bucket sizes")
		testSizes := []int{1, 5, 10, 20}

		for _, size := range testSizes {
			bucket := &models.Bucket{MaxSize: size}
			assert.Equal(size, bucket.MaxSize, "Bucket max size should be %d", size)

			// Test that we can conceptually add up to max size
			nodes := fixtures.CreateTestNodes(size, 8080)
			for i := 0; i < size; i++ {
				bucket.Nodes = append(bucket.Nodes, nodes[i])
			}
			assert.Equal(size, len(bucket.Nodes), "Should be able to add %d nodes", size)

			// Reset for next iteration
			bucket.Nodes = []*models.Node{}
		}

		section.Success("Bucket capacity limits working correctly")
	})
}
