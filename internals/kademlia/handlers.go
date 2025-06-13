package kademlia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Aradhya2708/kademlia/pkg/models"
)

// PingHandler handles /ping requests
func PingHandler(w http.ResponseWriter, r *http.Request, node *models.Node, routingTable *models.RoutingTable) {
	response := map[string]string{
		"message": "pong",
		"node_id": node.ID,
	}

	// Respond with a JSON object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FindNodeHandler handles /find_node requests
func FindNodeHandler(w http.ResponseWriter, r *http.Request, node *models.Node, routingTable *models.RoutingTable) {
	queryID := r.URL.Query().Get("id")
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
func StoreHandler(w http.ResponseWriter, r *http.Request, storage *models.KeyValueStore) {
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

	// Store the key-value pair using the thread-safe method
	storage.Set(kv.Key, kv.Value)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Stored key: %s, value: %s", kv.Key, kv.Value)
}

// FindValueHandler handles /find_value requests
func FindValueHandler(w http.ResponseWriter, r *http.Request, storage *models.KeyValueStore) {
	queryKey := r.URL.Query().Get("key")
	if queryKey == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	// Look up the value in storage
	if value, exists := storage.Store[queryKey]; exists {
		// Respond with the value
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	} else {
		// Key not found, respond with a 404
		http.Error(w, fmt.Sprintf("Key '%s' not found", queryKey), http.StatusNotFound)
	}
}
