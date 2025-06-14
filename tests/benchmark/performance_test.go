package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// BenchmarkStorageOperations benchmarks key-value storage operations
func BenchmarkStorageOperations(b *testing.B) {
	logger := testutils.NewTestLogger(nil, "BENCHMARK")
	fixtures := testutils.NewTestFixtures(logger)

	storage := kademlia.NewKeyValueStore()

	// Pre-generate a reasonable amount of test data
	numTestData := 1000
	if b.N > numTestData {
		numTestData = b.N
	}

	keys := make([]string, numTestData)
	values := make([]string, numTestData)
	for i := 0; i < numTestData; i++ {
		keys[i] = fixtures.GenerateValidHexID(fmt.Sprintf("bench%d", i))
		values[i] = fmt.Sprintf("benchmark-value-%d", i)
	}

	b.ResetTimer()

	// Benchmark Store operations
	b.Run("Store", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keyIndex := i % len(keys)
			kademlia.StoreKeyValue(storage, keys[keyIndex], values[keyIndex])
		}
	})

	// Store data for retrieval benchmark
	for i := 0; i < len(keys); i++ {
		kademlia.StoreKeyValue(storage, keys[i], values[i])
	}

	// Benchmark Retrieve operations
	b.Run("Retrieve", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keyIndex := i % len(keys)
			_, _ = kademlia.FindValue(storage, keys[keyIndex])
		}
	})
}

// BenchmarkRoutingTable benchmarks routing table operations
func BenchmarkRoutingTable(b *testing.B) {
	logger := testutils.NewTestLogger(nil, "BENCHMARK")
	fixtures := testutils.NewTestFixtures(logger)

	node := fixtures.CreateTestNode(8080, "bench")
	routingTable := kademlia.NewRoutingTable(node.ID)

	// Pre-generate a reasonable amount of test nodes
	numTestNodes := 100
	if b.N > numTestNodes {
		numTestNodes = b.N
	}

	testNodes := fixtures.CreateTestNodes(numTestNodes, 8081)

	b.ResetTimer()

	// Benchmark AddNode operations
	b.Run("AddNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeIndex := i % len(testNodes)
			kademlia.AddNodeToRoutingTable(routingTable, testNodes[nodeIndex], node.ID)
		}
	})

	// Ensure nodes are added for find benchmark
	for _, testNode := range testNodes {
		kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
	}

	// Benchmark FindClosestNodes operations
	b.Run("FindClosestNodes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			targetID := fixtures.GenerateValidHexID(fmt.Sprintf("target%d", i))
			kademlia.FindClosestNodes(routingTable, targetID, node.ID)
		}
	})
}

// TestPerformanceRegression runs performance tests to detect regressions
func TestPerformanceRegression(t *testing.T) {
	logger := testutils.NewTestLogger(t, "PERFORMANCE")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting performance regression tests")

	t.Run("StoragePerformance", func(t *testing.T) {
		section := logger.Section("Storage Performance")

		storage := kademlia.NewKeyValueStore()
		numOps := 1000

		section.Step(1, "Measure storage operation latency")
		start := time.Now()
		for i := 0; i < numOps; i++ {
			key := fixtures.GenerateValidHexID(fmt.Sprintf("perf%d", i))
			value := fmt.Sprintf("performance-value-%d", i)
			kademlia.StoreKeyValue(storage, key, value)
		}
		storeDuration := time.Since(start)

		section.Step(2, "Measure retrieval operation latency")
		start = time.Now()
		for i := 0; i < numOps; i++ {
			key := fixtures.GenerateValidHexID(fmt.Sprintf("perf%d", i))
			_, _ = kademlia.FindValue(storage, key)
		}
		retrieveDuration := time.Since(start)

		section.Step(3, "Verify performance thresholds")
		storeLatency := storeDuration.Nanoseconds() / int64(numOps)
		retrieveLatency := retrieveDuration.Nanoseconds() / int64(numOps)

		// Performance thresholds (adjust based on expected performance)
		maxStoreLatency := int64(1000000)   // 1ms per operation
		maxRetrieveLatency := int64(500000) // 0.5ms per operation

		assert.True(storeLatency < maxStoreLatency,
			"Store latency (%d ns) should be under threshold (%d ns)",
			storeLatency, maxStoreLatency)

		assert.True(retrieveLatency < maxRetrieveLatency,
			"Retrieve latency (%d ns) should be under threshold (%d ns)",
			retrieveLatency, maxRetrieveLatency)

		section.Info("Store latency: %d ns/op, Retrieve latency: %d ns/op",
			storeLatency, retrieveLatency)

		section.Success("Storage performance within acceptable limits")
	})

	t.Run("RoutingTablePerformance", func(t *testing.T) {
		section := logger.Section("Routing Table Performance")

		node := fixtures.CreateTestNode(8080, "perftest")
		routingTable := kademlia.NewRoutingTable(node.ID)

		section.Step(1, "Add large number of nodes")
		numNodes := 1000
		testNodes := fixtures.CreateTestNodes(numNodes, 8081)

		start := time.Now()
		for _, testNode := range testNodes {
			kademlia.AddNodeToRoutingTable(routingTable, testNode, node.ID)
		}
		addDuration := time.Since(start)

		section.Step(2, "Measure find operation performance")
		numFinds := 100
		start = time.Now()
		for i := 0; i < numFinds; i++ {
			targetID := fixtures.GenerateValidHexID(fmt.Sprintf("findperf%d", i))
			kademlia.FindClosestNodes(routingTable, targetID, node.ID)
		}
		findDuration := time.Since(start)

		section.Step(3, "Verify performance thresholds")
		addLatency := addDuration.Nanoseconds() / int64(numNodes)
		findLatency := findDuration.Nanoseconds() / int64(numFinds)

		// Performance thresholds
		maxAddLatency := int64(1000000)  // 1ms per add operation
		maxFindLatency := int64(5000000) // 5ms per find operation

		assert.True(addLatency < maxAddLatency,
			"Add node latency (%d ns) should be under threshold (%d ns)",
			addLatency, maxAddLatency)

		assert.True(findLatency < maxFindLatency,
			"Find nodes latency (%d ns) should be under threshold (%d ns)",
			findLatency, maxFindLatency)

		section.Info("Add node latency: %d ns/op, Find nodes latency: %d ns/op",
			addLatency, findLatency)

		section.Success("Routing table performance within acceptable limits")
	})
}
