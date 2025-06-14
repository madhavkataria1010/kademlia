package unit

import (
	"testing"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/models"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestKademliaJoinNetwork tests network joining functionality
func TestKademliaJoinNetwork(t *testing.T) {
	logger := testutils.NewTestLogger(t, "KADEMLIA")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Kademlia join network tests")

	t.Run("SuccessfulJoin", func(t *testing.T) {
		section := logger.Section("Successful Network Join")

		section.Step(1, "Setup bootstrap node and mock server")
		bootstrapNode := fixtures.CreateTestNode(8080, "bootstrap")
		mockServer := testutils.NewMockServer(section, bootstrapNode)
		defer mockServer.Close()

		section.Step(2, "Setup joining node")
		joiningNode := fixtures.CreateTestNode(8081, "joining")
		routingTable := kademlia.NewRoutingTable(joiningNode.ID)

		section.Step(3, "Attempt to join network")
		err := kademlia.JoinNetwork(joiningNode, routingTable, mockServer.GetAddress())

		assert.NoError(err, "Join should succeed")

		section.Step(4, "Verify bootstrap node in routing table")
		closestNodes := kademlia.FindClosestNodes(routingTable, bootstrapNode.ID, joiningNode.ID)

		found := false
		for _, node := range closestNodes {
			if node.ID == bootstrapNode.ID {
				found = true
				section.Info("Bootstrap node found in routing table")
				break
			}
		}
		assert.True(found, "Bootstrap node should be in routing table")

		section.Success("Network join successful")
	})

	t.Run("FailedJoin", func(t *testing.T) {
		section := logger.Section("Failed Network Join")

		section.Step(1, "Setup joining node")
		joiningNode := fixtures.CreateTestNode(8082, "failing")
		routingTable := kademlia.NewRoutingTable(joiningNode.ID)

		section.Step(2, "Attempt join with invalid address")
		invalidAddress := "nonexistent:99999"
		err := kademlia.JoinNetwork(joiningNode, routingTable, invalidAddress)

		assert.HasError(err, "Join should fail with invalid address")
		section.Success("Network join properly failed")
	})

	t.Run("InvalidAddressFormats", func(t *testing.T) {
		section := logger.Section("Invalid Address Formats")

		joiningNode := fixtures.CreateTestNode(8083, "testing")
		routingTable := kademlia.NewRoutingTable(joiningNode.ID)

		invalidAddresses := []string{
			"invalid",
			"invalid:port",
			"127.0.0.1:invalid",
			":8080",
			"127.0.0.1:",
			"",
		}

		for i, addr := range invalidAddresses {
			section.Step(i+1, "Testing invalid address: "+addr)
			err := kademlia.JoinNetwork(joiningNode, routingTable, addr)
			assert.HasError(err, "Should fail for invalid address: %s", addr)
		}

		section.Success("All invalid addresses properly rejected")
	})

	t.Run("JoinNetworkResponseHandling", func(t *testing.T) {
		section := logger.Section("Join Network Response Handling")

		section.Step(1, "Setup mock server with custom response")
		bootstrapNode := fixtures.CreateTestNode(8084, "custom")
		mockServer := testutils.NewMockServer(section, bootstrapNode)
		defer mockServer.Close()

		// Test with valid response
		section.Step(2, "Test with valid response")
		joiningNode := fixtures.CreateTestNode(8085, "valid")
		routingTable := kademlia.NewRoutingTable(joiningNode.ID)

		err := kademlia.JoinNetwork(joiningNode, routingTable, mockServer.GetAddress())
		assert.NoError(err, "Should succeed with valid response")

		// Test with invalid response (empty node ID)
		section.Step(3, "Test with invalid response")
		mockServer.SetResponse("ping", map[string]interface{}{
			"message": "pong",
			"node_id": "", // Empty node ID should cause error
		})

		joiningNode2 := fixtures.CreateTestNode(8086, "invalid")
		routingTable2 := kademlia.NewRoutingTable(joiningNode2.ID)

		err = kademlia.JoinNetwork(joiningNode2, routingTable2, mockServer.GetAddress())
		assert.HasError(err, "Should fail with empty node ID in response")

		section.Success("Response handling working correctly")
	})
}

// TestKademliaIntegration tests integration between different Kademlia components
func TestKademliaIntegration(t *testing.T) {
	logger := testutils.NewTestLogger(t, "KADEMLIA")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Kademlia integration tests")

	t.Run("RoutingTableAndStorage", func(t *testing.T) {
		section := logger.Section("Routing Table and Storage Integration")

		section.Step(1, "Setup components")
		localNodeID := fixtures.GenerateValidHexID("local")
		routingTable := kademlia.NewRoutingTable(localNodeID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Add nodes to routing table")
		testNodes := fixtures.CreateTestNodes(5, 8080)
		for _, node := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, node, localNodeID)
		}

		section.Step(3, "Store data")
		testData := fixtures.GetTestKeyValuePairs()
		for key, value := range testData {
			kademlia.StoreKeyValue(storage, key, value)
		}

		section.Step(4, "Verify both components work together")
		// Find closest nodes
		targetID := fixtures.GenerateValidHexID("target")
		closestNodes := kademlia.FindClosestNodes(routingTable, targetID, localNodeID)
		assert.True(len(closestNodes) > 0, "Should find closest nodes")

		// Verify storage
		for key, expectedValue := range testData {
			actualValue, exists := kademlia.FindValue(storage, key)
			assert.True(exists, "Should find stored value")
			assert.Equal(expectedValue, actualValue, "Value should match")
		}

		section.Success("Integration working correctly")
	})

	t.Run("MultipleNodeInteraction", func(t *testing.T) {
		section := logger.Section("Multiple Node Interaction")

		section.Step(1, "Create multiple nodes with routing tables")
		nodes := fixtures.CreateTestNodes(3, 8080)
		routingTables := make([]*models.RoutingTable, len(nodes))

		for i, node := range nodes {
			routingTables[i] = kademlia.NewRoutingTable(node.ID)
		}

		section.Step(2, "Each node knows about others")
		for i, node := range nodes {
			for j, otherNode := range nodes {
				if i != j {
					kademlia.AddNodeToRoutingTable(routingTables[i], otherNode, node.ID)
				}
			}
		}

		section.Step(3, "Verify connectivity")
		for i, node := range nodes {
			for j, otherNode := range nodes {
				if i != j {
					closestNodes := kademlia.FindClosestNodes(routingTables[i], otherNode.ID, node.ID)
					found := false
					for _, foundNode := range closestNodes {
						if foundNode.ID == otherNode.ID {
							found = true
							break
						}
					}
					assert.True(found, "Node %d should know about node %d", i, j)
				}
			}
		}

		section.Success("Multiple node interaction working correctly")
	})
}
