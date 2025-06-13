package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

// TODO: use IP to generate a unique ID
func GenerateNodeID() string {
	rand.Seed(time.Now().UnixNano())
	randomData := fmt.Sprintf("%d-%d", rand.Int63(), time.Now().UnixNano())
	hash := sha1.Sum([]byte(randomData))
	return hex.EncodeToString(hash[:])
}
