package rpc_test

import "testing"
import "github.com/stretchr/testify/assert"

import . "github.com/russellchadwick/rpc"

func TestGetService(t *testing.T) {
	discovery, err := NewDiscovery()
	assert.Nil(t, err)
	nodes, err := discovery.GetService("consul")
	assert.Nil(t, err)
	assert.NotNil(t, nodes, "Nodes is nil")
}

func TestRegisterService(t *testing.T) {
	discovery, err := NewDiscovery()
	assert.Nil(t, err)
	err = discovery.RegisterService("test", 9595)
	assert.Nil(t, err)
}

func TestUnregisterService(t *testing.T) {
	discovery, err := NewDiscovery()
	assert.Nil(t, err)
	err = discovery.DeregisterService("test")
	assert.Nil(t, err)
}
