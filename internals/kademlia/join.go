package kademlia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Aradhya2708/kademlia/pkg/models"
)

func JoinNetwork(node *models.Node, routingTable *models.RoutingTable, bootstrapAddr string) error {
	// Parse IP and port from bootstrapAddr
	parts := strings.Split(bootstrapAddr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid bootstrap address format, expected <ip>:<port>")
	}
	ip := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid port in bootstrap address: %v", err)
	}

	// Send a ping request to the bootstrap node
	url := fmt.Sprintf("http://%s/ping", bootstrapAddr)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to join network: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response to get the bootstrap node's ID
	var response struct {
		Message string `json:"message"`
		NodeID  string `json:"node_id"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response from bootstrap node: %v", err)
	}

	// Ensure the response contains a valid NodeID
	if response.NodeID == "" {
		return fmt.Errorf("invalid response from bootstrap node: missing node ID")
	}

	// Add bootstrap node to the routing table
	bootstrapNode := &models.Node{
		ID:   response.NodeID,
		IP:   ip,
		Port: port,
	}
	AddNodeToRoutingTable(routingTable, bootstrapNode, node.ID)

	return nil
}
