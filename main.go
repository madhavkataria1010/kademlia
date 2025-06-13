package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Aradhya2708/kademlia/cmd"
	"github.com/Aradhya2708/kademlia/internals/kademlia"
	"github.com/Aradhya2708/kademlia/pkg/models"
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

	fmt.Println("Welcome to Kademlia Distributed Hash Table (DHT) Node!")

	// Initialize node, routing table, and storage
	node := cmd.InitializeNode(port)
	routingTable := kademlia.NewRoutingTable(node.ID)
	storage := kademlia.NewKeyValueStore()

	fmt.Printf("hi")

	// Add the current node to its own routing table
	selfNode := &models.Node{
		ID:   node.ID,
		IP:   "127.0.0.1", // Use local IP as the node's IP address
		Port: port,
	}
	kademlia.AddNodeToRoutingTable(routingTable, selfNode, node.ID)

	log.Printf("Node initialized: ID=%s, IP=%s, Port=%d\n", node.ID, "127.0.0.1", port)

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
	cmd.StartServer(node, routingTable, storage, port)
}
