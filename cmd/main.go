package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
)

func main() {
	// Parse CLI arguments for node configuration
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <port> [<bootstrap_ip:bootstrap_port>] ")
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil || port <= 0 || port > 65535 {
		log.Fatalf("Invalid port: %v", os.Args[1])
	}

	var bootstrapAddr string
	if len(os.Args) > 2 {
		bootstrapAddr = os.Args[2]
	}

	// Initialize node, routing table, and storage
	node := initializeNode(port)
	routingTable := kademlia.NewRoutingTable(node.ID)
	storage := kademlia.NewKeyValueStore()

	if bootstrapAddr == "" {
		log.Println("No bootstrap address provided. Running in standalone mode.")
		log.Printf("Node ID: %s, Port: %d\n", node.ID, port)
		log.Println("This node is the starting point of a new network.")
	} else {
		// If bootstrap address provided, join the network
		log.Printf("Attempting to join the network via bootstrap node: %s\n", bootstrapAddr)
		err := kademlia.JoinNetwork(node, routingTable, bootstrapAddr)
		if err != nil {
			log.Fatalf("Failed to join network: %v", err)
		}
		log.Println("Successfully joined the network.")
	}

	// Start the server for Kademlia RPCs
	log.Printf("Starting Kademlia node on port %d...\n", port)
	startServer(node, routingTable, storage, port)
}
