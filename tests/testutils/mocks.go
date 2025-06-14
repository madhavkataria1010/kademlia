package testutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/Aradhya2708/kademlia/pkg/models"
)

// MockServer provides a mock HTTP server for testing network operations
type MockServer struct {
	server    *httptest.Server
	logger    *TestLogger
	node      *models.Node
	responses map[string]interface{}
}

// NewMockServer creates a new mock server
func NewMockServer(logger *TestLogger, node *models.Node) *MockServer {
	mock := &MockServer{
		logger:    logger,
		node:      node,
		responses: make(map[string]interface{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", mock.handlePing)
	mux.HandleFunc("/find_node", mock.handleFindNode)
	mux.HandleFunc("/store", mock.handleStore)
	mux.HandleFunc("/find_value", mock.handleFindValue)

	mock.server = httptest.NewServer(mux)

	// Extract port from server URL
	parts := strings.Split(mock.server.URL, ":")
	if len(parts) >= 3 {
		if port, err := strconv.Atoi(parts[2]); err == nil {
			mock.node.Port = port
		}
	}

	logger.Info("Started mock server at %s for node %s...", mock.server.URL, node.ID[:8])
	return mock
}

// Close shuts down the mock server
func (m *MockServer) Close() {
	m.server.Close()
	m.logger.Info("Closed mock server for node %s...", m.node.ID[:8])
}

// GetAddress returns the server address in host:port format
func (m *MockServer) GetAddress() string {
	url := m.server.URL
	return strings.TrimPrefix(url, "http://")
}

// SetResponse sets a custom response for a specific endpoint
func (m *MockServer) SetResponse(endpoint string, response interface{}) {
	m.responses[endpoint] = response
	m.logger.Info("Set custom response for %s endpoint", endpoint)
}

// handlePing handles ping requests
func (m *MockServer) handlePing(w http.ResponseWriter, r *http.Request) {
	m.logger.Info("Mock server received ping request")

	if customResp, exists := m.responses["ping"]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customResp)
		return
	}

	response := map[string]interface{}{
		"message": "pong",
		"node_id": m.node.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleFindNode handles find_node requests
func (m *MockServer) handleFindNode(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("id")
	m.logger.Info("Mock server received find_node request for ID: %s...", queryID[:8])

	if customResp, exists := m.responses["find_node"]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customResp)
		return
	}

	// Return the mock node itself as closest
	nodes := []*models.Node{m.node}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// handleStore handles store requests
func (m *MockServer) handleStore(w http.ResponseWriter, r *http.Request) {
	m.logger.Info("Mock server received store request")

	if customResp, exists := m.responses["store"]; exists {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, customResp)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Stored successfully")
}

// handleFindValue handles find_value requests
func (m *MockServer) handleFindValue(w http.ResponseWriter, r *http.Request) {
	queryKey := r.URL.Query().Get("key")
	m.logger.Info("Mock server received find_value request for key: %s...", queryKey[:8])

	if customResp, exists := m.responses["find_value"]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customResp)
		return
	}

	// Return not found by default (return closest nodes)
	nodes := []*models.Node{m.node}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// NetworkErrorMockServer simulates network errors
type NetworkErrorMockServer struct {
	logger *TestLogger
}

// NewNetworkErrorMockServer creates a server that always returns errors
func NewNetworkErrorMockServer(logger *TestLogger) *NetworkErrorMockServer {
	return &NetworkErrorMockServer{logger: logger}
}

// GetAddress returns an invalid address to simulate network errors
func (n *NetworkErrorMockServer) GetAddress() string {
	return "invalid:99999"
}

// Close is a no-op for error mock
func (n *NetworkErrorMockServer) Close() {
	// No-op
}
