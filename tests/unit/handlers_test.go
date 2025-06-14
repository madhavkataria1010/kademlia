package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestPingHandler tests the ping handler
func TestPingHandler(t *testing.T) {
	logger := testutils.NewTestLogger(t, "HANDLERS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting ping handler tests")

	t.Run("BasicPing", func(t *testing.T) {
		section := logger.Section("Basic Ping")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Create ping request")
		req, err := http.NewRequest("GET", "/ping", nil)
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute ping handler")
		rr := httptest.NewRecorder()
		kademlia.PingHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 OK")

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(err, "Response should be valid JSON")
		assert.Equal("pong", response["message"], "Should return pong message")
		assert.Equal(node.ID, response["node_id"], "Should return correct node ID")

		section.Success("Basic ping working correctly")
	})

	t.Run("PingWithNodeInfo", func(t *testing.T) {
		section := logger.Section("Ping with Node Info")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Create ping request with node info")
		pingerID := fixtures.GenerateValidHexID("pinger")
		req, err := http.NewRequest("GET", "/ping?id="+pingerID+"&port=8081", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute ping handler")
		rr := httptest.NewRecorder()
		kademlia.PingHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 OK")

		section.Step(5, "Verify routing table update")
		closestNodes := kademlia.FindClosestNodes(routingTable, pingerID, node.ID)
		found := false
		for _, foundNode := range closestNodes {
			if foundNode.ID == pingerID {
				found = true
				break
			}
		}
		assert.True(found, "Pinger should be added to routing table")

		section.Success("Ping with node info working correctly")
	})

	t.Run("PingWithInvalidPort", func(t *testing.T) {
		section := logger.Section("Ping with Invalid Port")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Test invalid ports")
		invalidPorts := []string{"0", "-1", "99999", "abc", ""}

		for _, port := range invalidPorts {
			section.Step(3, "Testing invalid port: "+port)
			pingerID := fixtures.GenerateValidHexID("pinger")
			req, err := http.NewRequest("GET", "/ping?id="+pingerID+"&port="+port, nil)
			req.RemoteAddr = "127.0.0.1:12345"
			assert.NoError(err, "Request creation should not error")

			rr := httptest.NewRecorder()
			kademlia.PingHandler(rr, req, node, storage, routingTable)

			assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for invalid port: %s", port)
		}

		section.Success("Invalid ports properly rejected")
	})
}

// TestFindNodeHandler tests the find_node handler
func TestFindNodeHandler(t *testing.T) {
	logger := testutils.NewTestLogger(t, "HANDLERS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting find_node handler tests")

	t.Run("ValidFindNode", func(t *testing.T) {
		section := logger.Section("Valid Find Node")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)

		// Add some test nodes to routing table
		testNodes := fixtures.CreateTestNodes(3, 8081)
		for _, testNode := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
		}

		section.Step(2, "Create find_node request")
		queryID := fixtures.GenerateValidHexID("query")
		req, err := http.NewRequest("GET", "/find_node?id="+queryID, nil)
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute find_node handler")
		rr := httptest.NewRecorder()
		kademlia.FindNodeHandler(rr, req, node, routingTable)

		section.Step(4, "Verify response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 OK")

		var nodes []map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &nodes)
		assert.NoError(err, "Response should be valid JSON")

		section.Success("Valid find_node working correctly")
	})

	t.Run("InvalidFindNodeID", func(t *testing.T) {
		section := logger.Section("Invalid Find Node ID")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)

		section.Step(2, "Test invalid IDs")
		invalidIDs := fixtures.GenerateInvalidIDs()

		for desc, invalidID := range invalidIDs {
			section.Step(3, "Testing invalid ID: "+desc)
			req, err := http.NewRequest("GET", "/find_node?id="+invalidID, nil)
			assert.NoError(err, "Request creation should not error")

			rr := httptest.NewRecorder()
			kademlia.FindNodeHandler(rr, req, node, routingTable)

			assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for invalid ID: %s", desc)
		}

		section.Success("Invalid IDs properly rejected")
	})

	t.Run("MissingFindNodeID", func(t *testing.T) {
		section := logger.Section("Missing Find Node ID")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)

		section.Step(2, "Create request without ID parameter")
		req, err := http.NewRequest("GET", "/find_node", nil)
		assert.NoError(err, "Request creation should not error")

		rr := httptest.NewRecorder()
		kademlia.FindNodeHandler(rr, req, node, routingTable)

		section.Step(3, "Verify error response")
		assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for missing ID")

		section.Success("Missing ID properly handled")
	})
}

// TestStoreHandler tests the store handler
func TestStoreHandler(t *testing.T) {
	logger := testutils.NewTestLogger(t, "HANDLERS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting store handler tests")

	t.Run("ValidStore", func(t *testing.T) {
		section := logger.Section("Valid Store")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		// Add self to routing table to ensure we're among closest
		kademlia.AddNodeToRoutingTable(routingTable, node, node.ID)

		section.Step(2, "Create store request")
		storeData := map[string]string{
			"key":   fixtures.GenerateValidHexID("store"),
			"value": "test-store-value",
		}

		jsonData, err := json.Marshal(storeData)
		assert.NoError(err, "JSON marshal should not error")

		req, err := http.NewRequest("POST", "/store", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		assert.NoError(err, "Request creation should not error")

		section.Step(3, "Execute store handler")
		rr := httptest.NewRecorder()
		kademlia.StoreHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify storage")
		assert.Equal(http.StatusCreated, rr.Code, "Should return 201 Created")

		// Verify data was stored
		storedValue, exists := storage.Get(storeData["key"])
		assert.True(exists, "Key should exist in storage")
		assert.Equal(storeData["value"], storedValue, "Stored value should match")

		section.Success("Valid store working correctly")
	})

	t.Run("InvalidStoreMethod", func(t *testing.T) {
		section := logger.Section("Invalid Store Method")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Test invalid HTTP methods")
		methods := []string{"GET", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			section.Step(3, "Testing method: "+method)
			req, err := http.NewRequest(method, "/store", nil)
			assert.NoError(err, "Request creation should not error")

			rr := httptest.NewRecorder()
			kademlia.StoreHandler(rr, req, node, storage, routingTable)

			assert.Equal(http.StatusMethodNotAllowed, rr.Code, "Should return 405 for method: %s", method)
		}

		section.Success("Invalid methods properly rejected")
	})

	t.Run("InvalidStoreJSON", func(t *testing.T) {
		section := logger.Section("Invalid Store JSON")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		section.Step(2, "Test invalid JSON")
		req, err := http.NewRequest("POST", "/store", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		assert.NoError(err, "Request creation should not error")

		rr := httptest.NewRecorder()
		kademlia.StoreHandler(rr, req, node, storage, routingTable)
		assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for invalid JSON")

		section.Step(3, "Test missing key")
		incompleteData := map[string]string{"value": "test"}
		jsonData, _ := json.Marshal(incompleteData)

		req, _ = http.NewRequest("POST", "/store", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr = httptest.NewRecorder()
		kademlia.StoreHandler(rr, req, node, storage, routingTable)
		assert.Equal(http.StatusBadRequest, rr.Code, "Should return 400 for missing key")

		section.Success("Invalid JSON properly handled")
	})

	t.Run("StoreNotClosestNode", func(t *testing.T) {
		section := logger.Section("Store Not Closest Node")

		section.Step(1, "Setup test components")
		node := fixtures.CreateTestNode(8080, "test")
		routingTable := kademlia.NewRoutingTable(node.ID)
		storage := kademlia.NewKeyValueStore()

		// Add other nodes but not self to routing table
		testNodes := fixtures.CreateTestNodes(3, 8081)
		for _, testNode := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
		}

		section.Step(2, "Create store request")
		storeData := map[string]string{
			"key":   fixtures.GenerateValidHexID("store"),
			"value": "test-store-value",
		}

		jsonData, _ := json.Marshal(storeData)
		req, _ := http.NewRequest("POST", "/store", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		section.Step(3, "Execute store handler")
		rr := httptest.NewRecorder()
		kademlia.StoreHandler(rr, req, node, storage, routingTable)

		section.Step(4, "Verify redirect response")
		assert.Equal(http.StatusOK, rr.Code, "Should return 200 with closest nodes")

		// Should return closest nodes instead of storing
		var closestNodes []map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &closestNodes)
		assert.NoError(err, "Response should be valid JSON")

		// Verify data was NOT stored
		_, exists := storage.Get(storeData["key"])
		assert.False(exists, "Key should not be stored when not closest")

		section.Success("Not closest node behavior working correctly")
	})
}
