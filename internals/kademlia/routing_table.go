package kademlia

import (
	"math/big"
	"sort"
	"strings"

	"github.com/Aradhya2708/kademlia/pkg/constants"
	"github.com/Aradhya2708/kademlia/pkg/models"
)

// NodeDistance represents a node along with its calculated distance.
type NodeDistance struct {
	Node     *models.Node
	Distance *big.Int
}

func NewRoutingTable(nodeID string) *models.RoutingTable {
	// Create a routing table with buckets for each bit of the node ID
	buckets := make([]*models.Bucket, len(nodeID)*4) // Assuming hex (4 bits per char) // TODO: Check if this is correct

	k := constants.GetK() // Get the default bucket size (k)

	for i := range buckets {
		buckets[i] = &models.Bucket{MaxSize: k} // Default bucket size (k)
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

	// TODO: Torrentium, Add a Trust Score. 

	// Add node if bucket is not full
	if len(bucket.Nodes) < bucket.MaxSize {
		bucket.Nodes = append(bucket.Nodes, target)
	} else {
		// Handle full bucket (eviction or ignore)
		bucket.Nodes = bucket.Nodes[1:] // Simplified eviction (FIFO)
		bucket.Nodes = append(bucket.Nodes, target)
	}
}

// FindClosestNodes retrieves the closest nodes to the given queryID.
func FindClosestNodes(routingTable *models.RoutingTable, queryID, localID string) []*models.Node {
	// Calculate the XOR distance and collect all nodes.
	var distances []NodeDistance

	for _, bucket := range routingTable.Buckets {
		for _, node := range bucket.Nodes {
			distance := calculateXORDistance(queryID, node.ID)
			distances = append(distances, NodeDistance{
				Node:     node,
				Distance: distance,
			})
		}
	}

	// Sort nodes by distance.
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].Distance.Cmp(distances[j].Distance) < 0
	})

	// Return up to k closest nodes.
	k := constants.GetK()

	closestNodes := make([]*models.Node, 0, k)
	for i := 0; i < len(distances) && i < k; i++ {
		closestNodes = append(closestNodes, distances[i].Node)
	}

	return closestNodes
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
