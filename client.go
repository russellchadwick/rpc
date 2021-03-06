package rpc

import (
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
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

	address, err := discovery.GetRandomServiceAddress(name)
	if err != nil {
		log.Fatalf("failed to get service %s with discovery service: %v", name, err)
		return nil, err
	}

	clientConn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, err
	}

	return clientConn, nil
}
