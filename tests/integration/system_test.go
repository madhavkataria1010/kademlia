package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/constants"
	"github.com/Aradhya2708/kademlia/pkg/models"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestFullKademliaWorkflow tests the complete Kademlia workflow
func TestFullKademliaWorkflow(t *testing.T) {
	logger := testutils.NewTestLogger(t, "INTEGRATION")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting full Kademlia workflow integration test")

	t.Run("MultiNodeNetwork", func(t *testing.T) {
		section := logger.Section("Multi-Node Network")

		// Set k to a reasonable value for testing
		originalK := constants.GetK()
		constants.SetK(3)
		defer constants.SetK(originalK)

		section.Step(1, "Initialize multiple nodes")
		numNodes := 5
		nodes := make([]*models.Node, numNodes)
		routingTables := make([]*models.RoutingTable, numNodes)
		storages := make([]*models.KeyValueStore, numNodes)
		servers := make([]*httptest.Server, numNodes)

		// Create nodes
		for i := 0; i < numNodes; i++ {
			nodes[i] = fixtures.CreateTestNode(8080+i, fmt.Sprintf("node%d", i))
			routingTables[i] = kademlia.NewRoutingTable(nodes[i].ID)
			storages[i] = kademlia.NewKeyValueStore()

			// Add self to routing table
			kademlia.AddNodeToRoutingTable(routingTables[i], nodes[i], nodes[i].ID)

			// Create HTTP server for each node
			servers[i] = createTestServer(nodes[i], routingTables[i], storages[i])
			section.Info("Created node %d: %s... on server %s",
				i, nodes[i].ID[:8], servers[i].URL)
		}

		// Cleanup
		defer func() {
			for _, server := range servers {
				server.Close()
			}
		}()

		section.Step(2, "Bootstrap network - nodes discover each other")
		// First node is bootstrap, others join through it
		for i := 1; i < numNodes; i++ {
			bootstrapAddr := getServerAddress(servers[0])
			err := kademlia.JoinNetwork(nodes[i], routingTables[i], bootstrapAddr)
			assert.NoError(err, "Node %d should join network successfully", i)
		}

		section.Step(3, "Verify network connectivity")
		time.Sleep(100 * time.Millisecond) // Allow routing tables to update

		for i := 0; i < numNodes; i++ {
			closestNodes := kademlia.FindClosestNodes(routingTables[i], nodes[0].ID, nodes[i].ID)
			assert.True(len(closestNodes) > 0, "Node %d should know about other nodes", i)
			section.Info("Node %d knows about %d other nodes", i, len(closestNodes))
		}

		section.Step(4, "Test distributed storage and retrieval")
		testData := map[string]string{
			fixtures.GenerateValidHexID("key1"): "distributed-value-1",
			fixtures.GenerateValidHexID("key2"): "distributed-value-2",
			fixtures.GenerateValidHexID("key3"): "distributed-value-3",
		}

		// Store data through different nodes
		keyIndex := 0
		for key, value := range testData {
			nodeIndex := keyIndex % numNodes
			success := storeValueOnNode(servers[nodeIndex], key, value)
			assert.True(success, "Should store key %s... on node %d", key[:8], nodeIndex)
			keyIndex++
		}

		section.Step(5, "Verify data can be retrieved from any node")
		for key, expectedValue := range testData {
			for i := 0; i < numNodes; i++ {
				actualValue, found := findValueOnNode(servers[i], key)
				if found {
					assert.Equal(expectedValue, actualValue,
						"Node %d should return correct value for key %s...", i, key[:8])
					section.Info("Node %d found key %s...", i, key[:8])
					break
				}
			}
		}

		section.Success("Multi-node network integration test completed")
	})
}

// TestNetworkResilience tests network behavior under various conditions
func TestNetworkResilience(t *testing.T) {
	logger := testutils.NewTestLogger(t, "INTEGRATION")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting network resilience tests")

	t.Run("NodeFailureRecovery", func(t *testing.T) {
		section := logger.Section("Node Failure Recovery")

		section.Step(1, "Setup 3-node network")
		nodes := make([]*models.Node, 3)
		routingTables := make([]*models.RoutingTable, 3)
		storages := make([]*models.KeyValueStore, 3)
		servers := make([]*httptest.Server, 3)

		for i := 0; i < 3; i++ {
			nodes[i] = fixtures.CreateTestNode(8080+i, fmt.Sprintf("resilient%d", i))
			routingTables[i] = kademlia.NewRoutingTable(nodes[i].ID)
			storages[i] = kademlia.NewKeyValueStore()
			kademlia.AddNodeToRoutingTable(routingTables[i], nodes[i], nodes[i].ID)
			servers[i] = createTestServer(nodes[i], routingTables[i], storages[i])
		}

		defer func() {
			for _, server := range servers {
				if server != nil {
					server.Close()
				}
			}
		}()

		section.Step(2, "Bootstrap network")
		for i := 1; i < 3; i++ {
			kademlia.JoinNetwork(nodes[i], routingTables[i], getServerAddress(servers[0]))
		}

		section.Step(3, "Store data before failure")
		testKey := fixtures.GenerateValidHexID("resilience")
		testValue := "resilience-test-value"
		success := storeValueOnNode(servers[0], testKey, testValue)
		assert.True(success, "Should store test data")

		section.Step(4, "Simulate node failure")
		servers[0].Close()
		servers[0] = nil
		section.Info("Shut down bootstrap node")

		section.Step(5, "Verify remaining nodes still functional")
		for i := 1; i < 3; i++ {
			pingResponse := pingNode(servers[i])
			assert.True(pingResponse, "Remaining nodes should still respond to ping")
		}

		section.Success("Network resilience test completed")
	})
}

// TestScalability tests system performance with varying loads
func TestScalability(t *testing.T) {
	logger := testutils.NewTestLogger(t, "INTEGRATION")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting scalability tests")

	t.Run("HighVolumeOperations", func(t *testing.T) {
		section := logger.Section("High Volume Operations")

		section.Step(1, "Setup test node")
		node := fixtures.CreateTestNode(8080, "scale")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()
		kademlia.AddNodeToRoutingTable(routingTable, node, node.ID)

		server := createTestServer(node, routingTable, storage)
		defer server.Close()

		section.Step(2, "Perform high volume storage operations")
		numOperations := 1000
		successCount := 0

		start := time.Now()
		for i := 0; i < numOperations; i++ {
			key := fixtures.GenerateValidHexID(fmt.Sprintf("scale%d", i))
			value := fmt.Sprintf("scale-value-%d", i)

			if storeValueOnNode(server, key, value) {
				successCount++
			}
		}
		duration := time.Since(start)

		section.Step(3, "Verify performance metrics")
		assert.Equal(numOperations, successCount, "All operations should succeed")

		opsPerSecond := float64(numOperations) / duration.Seconds()
		section.Info("Completed %d operations in %v (%.2f ops/sec)",
			numOperations, duration, opsPerSecond)

		assert.True(opsPerSecond > 100, "Should achieve reasonable throughput")

		section.Success("High volume operations completed")
	})

	t.Run("LargeRoutingTable", func(t *testing.T) {
		section := logger.Section("Large Routing Table")

		section.Step(1, "Setup node with large routing table")
		node := fixtures.CreateTestNode(8080, "large")
		routingTable := kademlia.NewRoutingTable(node.ID)

		section.Step(2, "Add many nodes to routing table")
		numNodes := 1000
		testNodes := fixtures.CreateTestNodes(numNodes, 8081)

		start := time.Now()
		for _, testNode := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
		}
		addDuration := time.Since(start)

		section.Step(3, "Test find operations on large table")
		targetID := fixtures.GenerateValidHexID("target")

		start = time.Now()
		closestNodes := kademlia.FindClosestNodes(routingTable, targetID, node.ID)
		findDuration := time.Since(start)

		section.Step(4, "Verify performance")
		k := constants.GetK()
		assert.True(len(closestNodes) <= k, "Should return at most k nodes")
		assert.True(addDuration < time.Second, "Adding nodes should be fast")
		assert.True(findDuration < 100*time.Millisecond, "Finding nodes should be fast")

		section.Info("Added %d nodes in %v, found closest in %v",
			numNodes, addDuration, findDuration)

		section.Success("Large routing table test completed")
	})
}

// Helper functions for integration tests

func createTestServer(node *models.Node, routingTable *models.RoutingTable, storage *models.KeyValueStore) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		kademlia.PingHandler(w, r, node, storage, routingTable)
	})
	mux.HandleFunc("/find_node", func(w http.ResponseWriter, r *http.Request) {
		kademlia.FindNodeHandler(w, r, node, routingTable)
	})
	mux.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		kademlia.StoreHandler(w, r, node, storage, routingTable)
	})
	mux.HandleFunc("/find_value", func(w http.ResponseWriter, r *http.Request) {
		kademlia.FindValueHandler(w, r, node, storage, routingTable)
	})

	return httptest.NewServer(mux)
}

func getServerAddress(server *httptest.Server) string {
	// Extract address from server URL (remove http://)
	url := server.URL
	return url[7:] // Remove "http://"
}

func storeValueOnNode(server *httptest.Server, key, value string) bool {
	storeData := map[string]string{"key": key, "value": value}
	jsonData, _ := json.Marshal(storeData)

	resp, err := http.Post(server.URL+"/store", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusCreated
}

func findValueOnNode(server *httptest.Server, key string) (string, bool) {
	resp, err := http.Get(server.URL + "/find_value?key=" + key)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	var value string
	if json.NewDecoder(resp.Body).Decode(&value) != nil {
		return "", false
	}

	return value, true
}

func pingNode(server *httptest.Server) bool {
	if server == nil {
		return false
	}

	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
