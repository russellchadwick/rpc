package rpc

import (
	"math/rand"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

// Client provides an RPC client that will find services by name
type Client struct {
}

// Dial takes the name of a service and figures out who to dial
func (c *Client) Dial(name string) (*grpc.ClientConn, error) {
	discovery, err := NewDiscovery()
	if err != nil {
		log.Fatalf("failed to create discovery service: %v", err)
		return nil, err
	}

	nodes, err := discovery.GetService(name)
	if err != nil {
		log.Fatalf("failed to get service %s with discovery service: %v", name, err)
		return nil, err
	}

	randomInt := rand.Intn(len(nodes))
	node := nodes[randomInt]
	address := net.JoinHostPort(node.Address, strconv.Itoa(node.Port))
	log.WithFields(log.Fields{
		"total":        len(nodes),
		"index_chosen": randomInt,
		"address":      address,
	}).Info("randomly chose node")

	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, err
	}

	return clientConn, nil
}
