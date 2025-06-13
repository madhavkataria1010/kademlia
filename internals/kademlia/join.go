package kademlia

import (
	"fmt"
	"net/http"

	"github.com/Aradhya2708/kademlia/pkg/models"
)

func JoinNetwork(node *models.Node, routingTable *models.RoutingTable, bootstrapAddr string) error {
	url := fmt.Sprintf("http://%s/ping", bootstrapAddr) // ping
	resp, err := http.Get(url) // wait for pong
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to join network: %v", err)
	}

	// Add bootstrap node to routing table
	bootstrapNode := &models.Node{
		ID:   "", // Ideally get this from the bootstrap response
		IP:   bootstrapAddr,
		Port: 0, // Parse port from bootstrapAddr
	}
	AddNodeToRoutingTable(routingTable, bootstrapNode, node.ID)
	return nil
}
