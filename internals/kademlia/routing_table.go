package kademlia

import (
	"math/big"
	"strings"

	"github.com/Aradhya2708/kademlia/pkg/models"
)

func NewRoutingTable(nodeID string) *models.RoutingTable {
	// Create a routing table with buckets for each bit of the node ID
	buckets := make([]*models.Bucket, len(nodeID)*4) // Assuming hex (4 bits per char)
	for i := range buckets {
		buckets[i] = &models.Bucket{MaxSize: 20} // Default bucket size (k)
	}
	return &models.RoutingTable{Buckets: buckets}
}

func AddNodeToRoutingTable(rt *models.RoutingTable, target *models.Node, localID string) {
	distance := calculateXORDistance(localID, target.ID)
	bucketIndex := getBucketIndex(distance)
	bucket := rt.Buckets[bucketIndex]

	// Ensure no duplicate entries
	for _, n := range bucket.Nodes {
		if n.ID == target.ID {
			return
		}
	}

	// Add node if bucket is not full
	if len(bucket.Nodes) < bucket.MaxSize {
		bucket.Nodes = append(bucket.Nodes, target)
	} else {
		// Handle full bucket (eviction or ignore)
		bucket.Nodes = bucket.Nodes[1:] // Simplified eviction (FIFO)
		bucket.Nodes = append(bucket.Nodes, target)
	}
}

func calculateXORDistance(id1, id2 string) *big.Int {
	bytes1 := decodeHex(id1)
	bytes2 := decodeHex(id2)
	xor := new(big.Int).Xor(new(big.Int).SetBytes(bytes1), new(big.Int).SetBytes(bytes2))
	return xor
}

func getBucketIndex(distance *big.Int) int {
	return distance.BitLen() - 1 // Most significant bit position
}

func decodeHex(s string) []byte {
	result, _ := new(big.Int).SetString(strings.ToUpper(s), 16)
	return result.Bytes()
}
