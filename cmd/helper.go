package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/models"
)

func InitializeNode(port int) *models.Node {

	fmt.Println("Initializing Kademlia node...")
	// Generate Node ID
	nodeID := kademlia.GenerateNodeID()

	// Create Node
	node := &models.Node{
		ID:   nodeID,
		IP:   "127.0.0.1", // Assuming localhost for now
		Port: port,
	}

	log.Printf("Initialized node: ID=%s, IP=%s, Port=%d\n", node.ID, node.IP, node.Port)
	return node
}

func StartServer(node *models.Node, routingTable *models.RoutingTable, storage *models.KeyValueStore, port int) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		kademlia.PingHandler(w, r, node, storage, routingTable)
	})
	http.HandleFunc("/find_node", func(w http.ResponseWriter, r *http.Request) {
		kademlia.FindNodeHandler(w, r, node, routingTable)
	})
	http.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		kademlia.StoreHandler(w, r, node, storage, routingTable)
	})
	http.HandleFunc("/find_value", func(w http.ResponseWriter, r *http.Request) {
		kademlia.FindValueHandler(w, r, node, storage, routingTable)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
