package unit

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Aradhya2708/kademlia/pkg/models"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestKeyValueStore tests the KeyValueStore model
func TestKeyValueStore(t *testing.T) {
	logger := testutils.NewTestLogger(t, "MODELS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting KeyValueStore tests")

	t.Run("KeyValueStoreCreation", func(t *testing.T) {
		section := logger.Section("KeyValueStore Creation")

		section.Step(1, "Create new store")
		store := models.NewKeyValueStore()

		assert.NotNil(store, "Store should not be nil")
		assert.NotNil(store.Store, "Internal store should not be nil")

		section.Step(2, "Verify empty store")
		allData := store.GetAll()
		assert.Equal(0, len(allData), "Store should start empty")

		section.Success("KeyValueStore created successfully")
	})

	t.Run("KeyValueStoreOperations", func(t *testing.T) {
		section := logger.Section("KeyValueStore Operations")

		section.Step(1, "Create store and test data")
		store := models.NewKeyValueStore()
		testData := fixtures.GetTestKeyValuePairs()

		section.Step(2, "Set values")
		for key, value := range testData {
			store.Set(key, value)
			section.Info("Set key %s... = %s", key[:8], value)
		}

		section.Step(3, "Get values")
		for key, expectedValue := range testData {
			actualValue, exists := store.Get(key)
			assert.True(exists, "Key %s... should exist", key[:8])
			assert.Equal(expectedValue, actualValue, "Value should match for key %s...", key[:8])
		}

		section.Step(4, "Get non-existent value")
		nonExistentKey := fixtures.GenerateValidHexID("nonexist")
		_, exists := store.Get(nonExistentKey)
		assert.False(exists, "Non-existent key should not exist")

		section.Step(5, "Get all values")
		allData := store.GetAll()
		assert.Equal(len(testData), len(allData), "GetAll should return all data")

		section.Success("KeyValueStore operations working correctly")
	})

	t.Run("KeyValueStoreConcurrency", func(t *testing.T) {
		section := logger.Section("KeyValueStore Concurrency")

		section.Step(1, "Create store for concurrency test")
		store := models.NewKeyValueStore()

		section.Step(2, "Concurrent operations test")
		var wg sync.WaitGroup
		numGoroutines := 10
		operationsPerGoroutine := 100

		// Concurrent writes
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fixtures.GenerateValidHexID(fmt.Sprintf("g%dop%d", goroutineID, j))
					value := fmt.Sprintf("value-%d-%d", goroutineID, j)
					store.Set(key, value)
				}
			}(i)
		}

		// Concurrent reads
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fixtures.GenerateValidHexID(fmt.Sprintf("g%dop%d", goroutineID, j))
					store.Get(key) // We don't check the result since timing is unpredictable
				}
			}(i)
		}

		wg.Wait()

		section.Step(3, "Verify final state")
		allData := store.GetAll()
		expectedCount := numGoroutines * operationsPerGoroutine
		assert.Equal(expectedCount, len(allData), "Should have all written values")

		section.Success("Concurrent operations completed successfully")
	})

	t.Run("KeyValueStoreEdgeCases", func(t *testing.T) {
		section := logger.Section("KeyValueStore Edge Cases")

		section.Step(1, "Test empty key and value")
		store := models.NewKeyValueStore()

		// Test empty strings (should be allowed)
		store.Set("", "")
		value, exists := store.Get("")
		assert.True(exists, "Empty key should be allowed")
		assert.Equal("", value, "Empty value should be stored")

		section.Step(2, "Test overwriting values")
		key := fixtures.GenerateValidHexID("overwrite")
		store.Set(key, "original")
		store.Set(key, "updated")

		value, exists = store.Get(key)
		assert.True(exists, "Key should still exist after overwrite")
		assert.Equal("updated", value, "Value should be updated")

		section.Step(3, "Test large values")
		largeValue := make([]byte, 10000)
		for i := range largeValue {
			largeValue[i] = 'A'
		}
		largeKey := fixtures.GenerateValidHexID("large")
		store.Set(largeKey, string(largeValue))

		retrievedValue, exists := store.Get(largeKey)
		assert.True(exists, "Large value should be stored")
		assert.Equal(len(largeValue), len(retrievedValue), "Large value should be retrieved correctly")

		section.Success("Edge cases handled correctly")
	})
}

// TestRoutingTableModel tests the RoutingTable model
func TestRoutingTableModel(t *testing.T) {
	logger := testutils.NewTestLogger(t, "MODELS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting RoutingTable model tests")

	t.Run("RoutingTableCreation", func(t *testing.T) {
		section := logger.Section("RoutingTable Creation")

		section.Step(1, "Create routing table")
		nodeID := fixtures.GenerateValidHexID("test")
		routingTable := &models.RoutingTable{}

		// Initialize buckets manually to test the model structure
		numBuckets := len(nodeID) * 4 // 40 chars * 4 bits = 160 buckets
		routingTable.Buckets = make([]*models.Bucket, numBuckets)

		for i := range routingTable.Buckets {
			routingTable.Buckets[i] = &models.Bucket{MaxSize: 1} // Use k=1 for testing
		}

		assert.NotNil(routingTable, "Routing table should not be nil")
		assert.Equal(numBuckets, len(routingTable.Buckets), "Should have correct number of buckets")

		section.Step(2, "Verify bucket initialization")
		for i, bucket := range routingTable.Buckets {
			assert.NotNil(bucket, "Bucket %d should not be nil", i)
			assert.Equal(1, bucket.MaxSize, "Bucket %d should have correct max size", i)
			assert.Equal(0, len(bucket.Nodes), "Bucket %d should start empty", i)
		}

		section.Success("RoutingTable created successfully")
	})

	t.Run("RoutingTableStructure", func(t *testing.T) {
		section := logger.Section("RoutingTable Structure")

		section.Step(1, "Test routing table with different node ID lengths")
		// Test with standard 40-character hex ID
		// No need to store the node ID since we're just testing table structure
		rt := &models.RoutingTable{}

		expectedBuckets := 40 * 4 // 160 buckets for 160-bit ID
		rt.Buckets = make([]*models.Bucket, expectedBuckets)

		assert.Equal(expectedBuckets, len(rt.Buckets), "Should have 160 buckets for 40-char hex ID")

		section.Step(2, "Verify bucket independence")
		// Initialize buckets and verify they're independent
		for i := range rt.Buckets {
			rt.Buckets[i] = &models.Bucket{MaxSize: i + 1} // Different sizes for testing
		}

		for i, bucket := range rt.Buckets {
			assert.Equal(i+1, bucket.MaxSize, "Bucket %d should have independent max size", i)
		}

		section.Success("RoutingTable structure working correctly")
	})
}

// TestMessageModel tests the Message model
func TestMessageModel(t *testing.T) {
	logger := testutils.NewTestLogger(t, "MODELS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting Message model tests")

	t.Run("MessageTypes", func(t *testing.T) {
		section := logger.Section("Message Types")

		section.Step(1, "Test all message type constants")
		messageTypes := []models.MessageType{
			models.Ping,
			models.FindNode,
			models.Store,
			models.FindValue,
			models.Pong,
		}

		expectedTypes := []string{"PING", "FIND_NODE", "STORE", "FIND_VALUE", "PONG"}

		for i, msgType := range messageTypes {
			assert.Equal(expectedTypes[i], string(msgType), "Message type %d should match expected", i)
		}

		section.Success("Message types defined correctly")
	})

	t.Run("MessageCreation", func(t *testing.T) {
		section := logger.Section("Message Creation")

		section.Step(1, "Create different message types")
		sender := fixtures.CreateTestNode(8080, "sender")

		// Test PING message
		pingMsg := models.Message{
			Type:   models.Ping,
			Sender: *sender,
		}
		assert.Equal(models.Ping, pingMsg.Type, "Ping message type should be correct")
		assert.Equal(sender.ID, pingMsg.Sender.ID, "Sender should be set correctly")

		// Test STORE message
		storeMsg := models.Message{
			Type:   models.Store,
			Sender: *sender,
			Key:    fixtures.GenerateValidHexID("storekey"),
			Value:  "test-value",
		}
		assert.Equal(models.Store, storeMsg.Type, "Store message type should be correct")
		assert.True(len(storeMsg.Key) == 40, "Store message key should be valid")
		assert.Equal("test-value", storeMsg.Value, "Store message value should be set")

		// Test FIND_NODE message
		findNodeMsg := models.Message{
			Type:   models.FindNode,
			Sender: *sender,
			Target: fixtures.GenerateValidHexID("target"),
		}
		assert.Equal(models.FindNode, findNodeMsg.Type, "FindNode message type should be correct")
		assert.True(len(findNodeMsg.Target) == 40, "FindNode message target should be valid")

		section.Success("Message creation working correctly")
	})
}
