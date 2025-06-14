package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/models"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestFindValueHandler tests the find_value handler
func TestFindValueHandler(t *testing.T) {
	logger := testutils.NewTestLogger(t, "HANDLERS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting find_value handler tests")

	t.Run("FindExistingValue", func(t *testing.T) {
		section := logger.Section("Find Existing Value")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		// Store test data
		testKey := fixtures.GenerateValidHexID("testkey")
		testValue := "test-value-find"
		storage.Set(testKey, testValue)

		section.Step(2, "Create find_value request")
		req, err := http.NewRequest("GET", "/find_value?key="+testKey, nil)
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute find_value handler")
		rr := httptest.NewRecorder()
		kademlia.FindValueHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 OK")

		var response string
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(err, "Response should be valid JSON")
		assert.Equal(testValue, response, "Should return the stored value")

		section.Success("Find existing value working correctly")
	})

	t.Run("FindNonExistentValue", func(t *testing.T) {
		section := logger.Section("Find Non-Existent Value")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		// Add some nodes to routing table for closest node response
		testNodes := fixtures.CreateTestNodes(3, 8081)
		for _, testNode := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
		}

		section.Step(2, "Create find_value request for non-existent key")
		nonExistentKey := fixtures.GenerateValidHexID("notexist")
		req, err := http.NewRequest("GET", "/find_value?key="+nonExistentKey, nil)
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute find_value handler")
		rr := httptest.NewRecorder()
		kademlia.FindValueHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 OK with closest nodes")

		// Should return closest nodes (like find_node)
		var closestNodes []map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &closestNodes)
		assert.NoError(err, "Response should be valid JSON array")

		section.Success("Find non-existent value working correctly")
	})

	t.Run("FindValueMissingKey", func(t *testing.T) {
		section := logger.Section("Find Value Missing Key")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Create request without key parameter")
		req, err := http.NewRequest("GET", "/find_value", nil)
		assert.NoError(err, "Request creation should not error")

		rr := httptest.NewRecorder()
		kademlia.FindValueHandler(rr, req, node, storage, routingTable)

		section.Step(3, "Verify error response")
		assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for missing key")

		section.Success("Missing key properly handled")
	})

	t.Run("FindValueInvalidKey", func(t *testing.T) {
		section := logger.Section("Find Value Invalid Key")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Test invalid keys")
		invalidKeys := fixtures.GenerateInvalidIDs()

		for desc, invalidKey := range invalidKeys {
			section.Step(3, "Testing invalid key: "+desc)
			req, err := http.NewRequest("GET", "/find_value?key="+invalidKey, nil)
			assert.NoError(err, "Request creation should not error")

			rr := httptest.NewRecorder()
			kademlia.FindValueHandler(rr, req, node, storage, routingTable)

			assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for invalid key: %s", desc)
		}

		section.Success("Invalid keys properly rejected")
	})
}

// TestHandlerIntegration tests integration between different handlers
func TestHandlerIntegration(t *testing.T) {
	logger := testutils.NewTestLogger(t, "HANDLERS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting handler integration tests")

	t.Run("PingAndFindNode", func(t *testing.T) {
		section := logger.Section("Ping and Find Node Integration")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Add node via ping")
		pingerID := fixtures.GenerateValidHexID("pinger")
		pingReq, _ := http.NewRequest("GET", "/ping?id="+pingerID+"&port=8081", nil)
		pingReq.RemoteAddr = "127.0.0.1:12345"

		pingRR := httptest.NewRecorder()
		kademlia.PingHandler(pingRR, pingReq, node, storage, routingTable)
		assert.Equal(http.StatusOK, pingRR.Code, "Ping should succeed")

		section.Step(3, "Find the added node")
		findReq, _ := http.NewRequest("GET", "/find_node?id="+pingerID, nil)
		findRR := httptest.NewRecorder()
		kademlia.FindNodeHandler(findRR, findReq, node, routingTable)

		assert.Equal(http.StatusOK, findRR.Code, "Find node should succeed")

		var foundNodes []map[string]interface{}
		json.Unmarshal(findRR.Body.Bytes(), &foundNodes)

		// Should find the pinged node
		found := false
		for _, foundNode := range foundNodes {
			if nodeID, ok := foundNode["ID"].(string); ok && nodeID == pingerID {
				found = true
				break
			}
		}
		assert.True(found, "Should find the node that was added via ping")

		section.Success("Ping and find node integration working correctly")
	})

	t.Run("StoreAndFindValue", func(t *testing.T) {
		section := logger.Section("Store and Find Value Integration")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		// Add self to routing table to ensure we can store
		kademlia.AddNodeToRoutingTable(routingTable, node, node.ID)

		section.Step(2, "Store a value")
		testKey := fixtures.GenerateValidHexID("integ")
		testValue := "integration-test-value"

		storeData := map[string]string{"key": testKey, "value": testValue}
		jsonData, _ := json.Marshal(storeData)

		storeReq, _ := http.NewRequest("POST", "/store", bytes.NewBuffer(jsonData))
		storeReq.Header.Set("Content-Type", "application/json")

		storeRR := httptest.NewRecorder()
		kademlia.StoreHandler(storeRR, storeReq, node, storage, routingTable)
		assert.Equal(http.StatusCreated, storeRR.Code, "Store should succeed")

		section.Step(3, "Find the stored value")
		findReq, _ := http.NewRequest("GET", "/find_value?key="+testKey, nil)
		findRR := httptest.NewRecorder()
		kademlia.FindValueHandler(findRR, findReq, node, storage, routingTable)

		assert.Equal(http.StatusOK, findRR.Code, "Find value should succeed")

		var foundValue string
		json.Unmarshal(findRR.Body.Bytes(), &foundValue)
		assert.Equal(testValue, foundValue, "Should find the stored value")

		section.Success("Store and find value integration working correctly")
	})

	t.Run("CompleteWorkflow", func(t *testing.T) {
		section := logger.Section("Complete Workflow")

		section.Step(1, "Setup multiple nodes")
		nodes := fixtures.CreateTestNodes(3, 8080)
		routingTables := make([]*models.RoutingTable, len(nodes))
		storages := make([]*models.KeyValueStore, len(nodes))

		for i, node := range nodes {
			routingTables[i] = kademlia.NewRoutingTable(node.ID)
			storages[i] = kademlia.NewKeyValueStore()
			// Add self to routing table
			kademlia.AddNodeToRoutingTable(routingTables[i], node, node.ID)
		}

		section.Step(2, "Nodes discover each other via ping")
		for i, node := range nodes {
			for j, otherNode := range nodes {
				if i != j {
					pingReq, _ := http.NewRequest("GET",
						fmt.Sprintf("/ping?id=%s&port=%d", otherNode.ID, otherNode.Port), nil)
					pingReq.RemoteAddr = fmt.Sprintf("%s:%d", otherNode.IP, otherNode.Port)

					pingRR := httptest.NewRecorder()
					kademlia.PingHandler(pingRR, pingReq, node, storages[i], routingTables[i])
					assert.Equal(http.StatusOK, pingRR.Code, "Ping should succeed")
				}
			}
		}

		section.Step(3, "Store data on appropriate nodes")
		testKey := fixtures.GenerateValidHexID("workflow")
		testValue := "workflow-test-value"

		// Try to store on first node
		storeData := map[string]string{"key": testKey, "value": testValue}
		jsonData, _ := json.Marshal(storeData)

		storeReq, _ := http.NewRequest("POST", "/store", bytes.NewBuffer(jsonData))
		storeReq.Header.Set("Content-Type", "application/json")

		storeRR := httptest.NewRecorder()
		kademlia.StoreHandler(storeRR, storeReq, nodes[0], storages[0], routingTables[0])

		// Should either store or redirect
		assert.True(storeRR.Code == http.StatusCreated || storeRR.Code == http.StatusOK,
			"Store should either succeed or redirect")

		section.Step(4, "Verify network can find the value")
		// Try to find from different nodes
		for i, node := range nodes {
			findReq, _ := http.NewRequest("GET", "/find_value?key="+testKey, nil)
			findRR := httptest.NewRecorder()
			kademlia.FindValueHandler(findRR, findReq, node, storages[i], routingTables[i])

			assert.Equal(http.StatusOK, findRR.Code, "Find value should succeed from node %d", i)
		}

		section.Success("Complete workflow integration working correctly")
	})
}

// BenchmarkHandlers benchmarks handler performance
func BenchmarkHandlers(b *testing.B) {
	// Create a dummy test for benchmark logging
	t := &testing.T{}
	logger := testutils.NewTestLogger(t, "BENCH")
	fixtures := testutils.NewTestFixtures(logger)

	// Setup
	node := fixtures.CreateTestNode(8080, "bench")
	routingTable := kademlia.NewRoutingTable(node.ID)
	storage := kademlia.NewKeyValueStore()

	logger.Info("Starting handler benchmarks")

	b.Run("PingHandler", func(b *testing.B) {
		req, _ := http.NewRequest("GET", "/ping", nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			kademlia.PingHandler(rr, req, node, storage, routingTable)
		}
	})

	b.Run("FindNodeHandler", func(b *testing.B) {
		queryID := fixtures.GenerateValidHexID("bench")
		req, _ := http.NewRequest("GET", "/find_node?id="+queryID, nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			kademlia.FindNodeHandler(rr, req, node, routingTable)
		}
	})

	logger.Info("Handler benchmarks completed")
}
