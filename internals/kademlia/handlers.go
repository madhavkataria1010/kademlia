package kademlia

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	validators "github.com/Aradhya2708/kademlia/internals/validator"
	"github.com/Aradhya2708/kademlia/pkg/models"
)

// PingHandler handles /ping requests
func PingHandler(w http.ResponseWriter, r *http.Request, node *models.Node, storage *models.KeyValueStore, routingTable *models.RoutingTable) {
	fmt.Println("Received ping request from:", r.RemoteAddr)

	// Extract pinger details from query parameters
	pingerID := r.URL.Query().Get("id")
	pingerPort := r.URL.Query().Get("port")

	if pingerID != "" && pingerPort != "" {
		// Pinger is a node, attempt to parse the port
		pingerUDPPort, err := strconv.Atoi(pingerPort)
		if err != nil || pingerUDPPort <= 0 || pingerUDPPort > 65535 {
			http.Error(w, "Invalid UDP port provided", http.StatusBadRequest)
			return
		}

		// Extract the IP address from the RemoteAddr
		pingerIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Failed to extract IP address", http.StatusInternalServerError)
			return
		}

		// Add the pinger node to the routing table
		pingerNode := &models.Node{
			ID:   pingerID,
			IP:   pingerIP,
			Port: pingerUDPPort,
		}
		AddNodeToRoutingTable(routingTable, pingerNode, node.ID)
		fmt.Printf("Added node to routing table: ID: %s, IP: %s, Port: %d\n", pingerID, pingerIP, pingerUDPPort)
	}

	// Debug: Print Current Node Details
	fmt.Println("Current Node Details:")
	fmt.Printf("ID: %s, IP: %s, Port: %d\n", node.ID, node.IP, node.Port)

	// Debug: Print Routing Table
	fmt.Println("Routing Table Details:")
	for i, bucket := range routingTable.Buckets {
		fmt.Printf("Bucket %d: ", i)
		for _, n := range bucket.Nodes {
			fmt.Printf("NodeID: %s, IP: %s, Port: %d | ", n.ID, n.IP, n.Port)
		}
		fmt.Println()
	}

	// Debug: Print Key-Value Store
	fmt.Println("Key-Value Store Contents:")
	for key, value := range storage.GetAll() {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}

	// Respond to the pinger
	response := map[string]interface{}{
		"message": "pong",
		"node_id": node.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FindNodeHandler handles /find_node requests
func FindNodeHandler(w http.ResponseWriter, r *http.Request, node *models.Node, routingTable *models.RoutingTable) {
	fmt.Println("Received ping find node req from:", r.RemoteAddr)

	queryID := r.URL.Query().Get("id")

	err := validators.ValidateID(queryID, validators.HexadecimalValidator)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid ID format: %v", err), http.StatusBadRequest)
		return
	}

	if queryID == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	// Find the closest nodes to the query ID
	closestNodes := FindClosestNodes(routingTable, queryID, node.ID)

	// Respond with the closest nodes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(closestNodes)
}

// StoreHandler handles /store requests
func StoreHandler(w http.ResponseWriter, r *http.Request, node *models.Node, storage *models.KeyValueStore, routingTable *models.RoutingTable) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Define a struct to parse incoming JSON
	var kv struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	// Read and parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &kv)
	if err != nil || kv.Key == "" || kv.Value == "" {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err = validators.ValidateID(kv.Key, validators.HexadecimalValidator)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid Key format: %v", err), http.StatusBadRequest)
		return
	}

	// Find the k closest nodes to the key
	closestNodes := FindClosestNodes(routingTable, kv.Key, node.ID)

	// // Calculate the XOR distance of this node to the key
	// ownDistance := calculateXORDistance(node.ID, kv.Key) ? why

	// Check if this node is among the k closest
	isClosest := false
	for _, peer := range closestNodes {
		if peer.ID == node.ID {
			isClosest = true
			break
		}
	}

	// If not among the closest, respond with the k closest nodes
	if !isClosest {
		fmt.Println("Node is not among the closest nodes, returning closest nodes")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(closestNodes)
		return
	}

	// Store the key-value pair if the node is among the closest
	storage.Set(kv.Key, kv.Value)
	fmt.Println("Stored key-value pair:", kv.Key, kv.Value)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Stored key: %s, value: %s", kv.Key, kv.Value)
}

// FindValueHandler handles /find_value requests
func FindValueHandler(w http.ResponseWriter, r *http.Request, node *models.Node, storage *models.KeyValueStore, routingTable *models.RoutingTable) {
	queryKey := r.URL.Query().Get("key")
	if queryKey == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	err := validators.ValidateID(queryKey, validators.HexadecimalValidator)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid Key format: %v", err), http.StatusBadRequest)
		return
	}

	// Look up the value in storage
	if value, exists := storage.Store[queryKey]; exists {
		// Respond with the value
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	} else {
		// Key not found, respond with a 404
		// http.Error(w, fmt.Sprintf("Key '%s' not found", queryKey), http.StatusNotFound)

		// key not found, respond as FIND_NODE res
		closestNodes := FindClosestNodes(routingTable, queryKey, node.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(closestNodes)
	}
}
