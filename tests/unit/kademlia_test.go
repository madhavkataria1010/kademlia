package unit

import (
	"math/big"
	"testing"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/constants"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestKademliaID tests ID generation and validation
func TestKademliaID(t *testing.T) {
	logger := testutils.NewTestLogger(t, "KADEMLIA")
	assert := testutils.NewAssert(logger)

	logger.Info("Starting Kademlia ID tests")

	t.Run("IDGeneration", func(t *testing.T) {
		section := logger.Section("ID Generation")

		section.Step(1, "Generate multiple IDs")
		ids := make([]string, 10)
		for i := 0; i < 10; i++ {
			ids[i] = kademlia.GenerateNodeID()
			section.Info("Generated ID %d: %s", i+1, ids[i])
		}

		section.Step(2, "Verify ID properties")
		for i, id := range ids {
			assert.Equal(40, len(id), "ID %d should be 40 characters long", i+1)
			assert.True(isHexString(id), "ID %d should be valid hex string", i+1)
		}

		section.Step(3, "Verify ID uniqueness")
		uniqueIDs := make(map[string]bool)
		for i, id := range ids {
			assert.False(uniqueIDs[id], "ID %d should be unique: %s", i+1, id)
			uniqueIDs[id] = true
		}

		section.Success("ID generation working correctly")
	})

	t.Run("IDRandomness", func(t *testing.T) {
		section := logger.Section("ID Randomness")

		section.Step(1, "Generate large number of IDs")
		numIDs := 100
		ids := make([]string, numIDs)

		for i := 0; i < numIDs; i++ {
			ids[i] = kademlia.GenerateNodeID()
		}

		section.Step(2, "Check for patterns")
		// Check that no two IDs are identical
		idSet := make(map[string]bool)
		duplicates := 0

		for _, id := range ids {
			if idSet[id] {
				duplicates++
			}
			idSet[id] = true
		}

		assert.Equal(0, duplicates, "Should have no duplicate IDs")
		assert.Equal(numIDs, len(idSet), "Should have generated %d unique IDs", numIDs)

		section.Success("ID randomness verified")
	})
}

// TestKademliaStorage tests storage operations
func TestKademliaStorage(t *testing.T) {
	logger := testutils.NewTestLogger(t, "KADEMLIA")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Kademlia storage tests")

	t.Run("StorageOperations", func(t *testing.T) {
		section := logger.Section("Storage Operations")

		section.Step(1, "Create storage instance")
		storage := kademlia.NewKeyValueStore()
		assert.NotNil(storage, "Storage should not be nil")

		section.Step(2, "Store and retrieve values")
		testData := fixtures.GetTestKeyValuePairs()

		for key, value := range testData {
			kademlia.StoreKeyValue(storage, key, value)
			section.Info("Stored: %s... = %s", key[:8], value)
		}

		section.Step(3, "Verify stored values")
		for key, expectedValue := range testData {
			actualValue, exists := kademlia.FindValue(storage, key)
			assert.True(exists, "Key should exist: %s...", key[:8])
			assert.Equal(expectedValue, actualValue, "Values should match for key: %s...", key[:8])
		}

		section.Step(4, "Test non-existent key")
		nonExistentKey := fixtures.GenerateValidHexID("nonexist")
		_, exists := kademlia.FindValue(storage, nonExistentKey)
		assert.False(exists, "Non-existent key should not be found")

		section.Success("Storage operations working correctly")
	})

	t.Run("StorageWrapperFunctions", func(t *testing.T) {
		section := logger.Section("Storage Wrapper Functions")

		section.Step(1, "Test StoreKeyValue wrapper")
		storage := kademlia.NewKeyValueStore()
		key := fixtures.GenerateValidHexID("wrapper")
		value := "wrapper-test-value"

		kademlia.StoreKeyValue(storage, key, value)

		// Verify using direct access
		directValue, exists := storage.Get(key)
		assert.True(exists, "Value should be stored via wrapper")
		assert.Equal(value, directValue, "Wrapper should store correct value")

		section.Step(2, "Test FindValue wrapper")
		wrapperValue, wrapperExists := kademlia.FindValue(storage, key)
		assert.True(wrapperExists, "Wrapper should find stored value")
		assert.Equal(value, wrapperValue, "Wrapper should return correct value")

		section.Success("Storage wrapper functions working correctly")
	})
}

// TestKademliaRoutingTable tests routing table operations
func TestKademliaRoutingTable(t *testing.T) {
	logger := testutils.NewTestLogger(t, "KADEMLIA")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Kademlia routing table tests")

	t.Run("RoutingTableCreation", func(t *testing.T) {
		section := logger.Section("Routing Table Creation")

		section.Step(1, "Create routing table")
		nodeID := fixtures.GenerateValidHexID("test")
		routingTable := kademlia.NewRoutingTable(nodeID)

		assert.NotNil(routingTable, "Routing table should not be nil")
		assert.True(len(routingTable.Buckets) > 0, "Routing table should have buckets")

		expectedBuckets := len(nodeID) * 4 // 40 chars * 4 bits = 160 buckets
		assert.Equal(expectedBuckets, len(routingTable.Buckets), "Should have correct number of buckets")

		section.Step(2, "Verify bucket initialization")
		k := constants.GetK()
		for i, bucket := range routingTable.Buckets {
			assert.NotNil(bucket, "Bucket %d should not be nil", i)
			assert.Equal(k, bucket.MaxSize, "Bucket %d should have max size k", i)
			assert.Equal(0, len(bucket.Nodes), "Bucket %d should start empty", i)
		}

		section.Success("Routing table created successfully")
	})

	t.Run("NodeAddition", func(t *testing.T) {
		section := logger.Section("Node Addition")

		section.Step(1, "Setup routing table and test nodes")
		localNodeID := fixtures.GenerateValidHexID("local")
		routingTable := kademlia.NewRoutingTable(localNodeID)
		testNodes := fixtures.CreateTestNodes(5, 8080)

		section.Step(2, "Add nodes to routing table")
		for i, node := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, node, localNodeID)
			section.Info("Added node %d: %s...", i+1, node.ID[:8])
		}

		section.Step(3, "Verify nodes were added")
		totalNodesFound := 0
		for _, node := range testNodes {
			closestNodes := kademlia.FindClosestNodes(routingTable, node.ID, localNodeID)
			for _, foundNode := range closestNodes {
				if foundNode.ID == node.ID {
					totalNodesFound++
					break
				}
			}
		}

		assert.True(totalNodesFound > 0, "At least some nodes should be found in routing table")
		section.Info("Found %d nodes in routing table", totalNodesFound)

		section.Success("Node addition working correctly")
	})

	t.Run("DuplicateNodePrevention", func(t *testing.T) {
		section := logger.Section("Duplicate Node Prevention")

		section.Step(1, "Setup routing table")
		localNodeID := fixtures.GenerateValidHexID("local")
		routingTable := kademlia.NewRoutingTable(localNodeID)
		testNode := fixtures.CreateTestNode(8080, "dup")

		section.Step(2, "Add same node multiple times")
		for i := 0; i < 5; i++ {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, localNodeID)
		}

		section.Step(3, "Verify only one instance exists")
		closestNodes := kademlia.FindClosestNodes(routingTable, testNode.ID, localNodeID)

		duplicateCount := 0
		for _, node := range closestNodes {
			if node.ID == testNode.ID {
				duplicateCount++
			}
		}

		assert.Equal(1, duplicateCount, "Should only have one instance of the node")

		section.Success("Duplicate node prevention working correctly")
	})

	t.Run("FindClosestNodes", func(t *testing.T) {
		section := logger.Section("Find Closest Nodes")

		// Temporarily set k to a reasonable value for testing
		originalK := constants.GetK()
		constants.SetK(3)
		defer constants.SetK(originalK)

		section.Step(1, "Setup routing table with nodes")
		localNodeID := fixtures.GenerateValidHexID("local")
		routingTable := kademlia.NewRoutingTable(localNodeID)

		// Add multiple test nodes
		testNodes := fixtures.CreateTestNodes(10, 8080)
		for _, node := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, node, localNodeID)
		}

		section.Step(2, "Find closest nodes to target")
		targetID := fixtures.GenerateValidHexID("target")
		closestNodes := kademlia.FindClosestNodes(routingTable, targetID, localNodeID)

		section.Step(3, "Verify results")
		k := constants.GetK()
		assert.True(len(closestNodes) <= k, "Should return at most k nodes")
		section.Info("Found %d closest nodes (k=%d)", len(closestNodes), k)

		// Verify nodes are actually closest (basic check)
		if len(closestNodes) > 1 {
			// Calculate distances and verify ordering
			distances := make([]*big.Int, len(closestNodes))
			for i, node := range closestNodes {
				distances[i] = calculateXORDistance(targetID, node.ID)
			}

			// Check if distances are in ascending order
			for i := 1; i < len(distances); i++ {
				assert.True(distances[i-1].Cmp(distances[i]) <= 0,
					"Nodes should be ordered by distance")
			}
		}

		section.Success("Find closest nodes working correctly")
	})

	t.Run("XORDistanceCalculation", func(t *testing.T) {
		section := logger.Section("XOR Distance Calculation")

		section.Step(1, "Test distance calculation")
		id1 := fixtures.GenerateValidHexID("test1")
		id2 := fixtures.GenerateValidHexID("test2")

		// Calculate distance using our helper function
		distance := calculateXORDistance(id1, id2)
		assert.NotNil(distance, "Distance should not be nil")
		assert.True(distance.Cmp(big.NewInt(0)) >= 0, "Distance should be non-negative")

		section.Step(2, "Test distance symmetry")
		distance1 := calculateXORDistance(id1, id2)
		distance2 := calculateXORDistance(id2, id1)
		assert.Equal(0, distance1.Cmp(distance2), "Distance should be symmetric")

		section.Step(3, "Test distance to self")
		distanceToSelf := calculateXORDistance(id1, id1)
		assert.Equal(0, distanceToSelf.Cmp(big.NewInt(0)), "Distance to self should be 0")

		section.Success("XOR distance calculation working correctly")
	})
}

// Helper function to check if string is valid hex
func isHexString(s string) bool {
	for _, char := range s {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

// Helper function to calculate XOR distance (duplicated for testing)
func calculateXORDistance(id1, id2 string) *big.Int {
	big1, _ := big.NewInt(0).SetString(id1, 16)
	big2, _ := big.NewInt(0).SetString(id2, 16)
	return big.NewInt(0).Xor(big1, big2)
}
