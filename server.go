package rpc

import (
	"net"

	"math/rand"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"net/http"
	// registers pprof http handlers
	_ "net/http/pprof"
)

// Server provides an RPC server that registers itself with discovery and uses HTTP/2 transport
type Server struct {
	name     string
	listener net.Listener
}

// Serve will register with discovery and wait for queries
func (s *Server) Serve(name string, registrationFunc func(*grpc.Server)) error {
	s.name = name

	port, err := findExistingPort(name)
	if err != nil {
		log.WithField("error", err).Fatal("error finding port")
		return err
	}

	if port == nil {
		port = randomPort()
	}

	s.listener, err = net.Listen("tcp", laddr(*port))
	if err != nil {
		log.WithField("port", port).WithField("error", err).Fatal("failed to listen")
		return err
	}

	grpcServer := grpc.NewServer()
	grpc.EnableTracing = true
	registrationFunc(grpcServer)
	go grpcServer.Serve(s.listener)

	err = registerWithDiscovery(name, *port)
	if err != nil {
		log.WithField("error", err).Error("failed to register with discovery")
		return err
	}

	http.Handle("/metrics", prometheus.Handler())
	err = http.Serve(s.listener, nil)
	if err != nil {
		log.WithField("error", err).Error("failed to serve http")
		return err
	}

	return nil
}

// Stop will end serving and remove itself from discovery
func (s *Server) Stop() error {
	err := s.listener.Close()
	if err != nil {
		log.WithField("error", err).Error("failed to stop listener")
		return err
	}

	err = deregisterWithDiscovery(s.name)
	if err != nil {
		log.WithField("error", err).Error("failed to deregister with discovery")
		return err
	}

	return nil
}

func findExistingPort(name string) (*int, error) {
	discovery, err := connectToDiscovery()
	services, err := discovery.GetLocalServices()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if strings.EqualFold(service.Name, name) {
			log.WithField("port", service.Node.Port).Debug("found existing port")
			return &service.Node.Port, nil
		}
	}

	return nil, nil
}

func randomPort() *int {
	rand.Seed(time.Now().UTC().UnixNano())
	port := rand.Intn(1000) + 50000
	log.WithField("port", port).Debug("Chose random port")
	return &port
}

func laddr(port int) string {
	return ":" + strconv.Itoa(port)
}

func registerWithDiscovery(name string, port int) error {
	discovery, err := connectToDiscovery()
	if err != nil {
		return err
	}

	err = discovery.RegisterService(name, port)
	if err != nil {
		log.Fatalf("failed to register with discovery service: %v", err)
		return err
	}

	log.WithField("name", name).WithField("port", port).Info("registered with discovery service")

	return nil
}

func deregisterWithDiscovery(name string) error {
	discovery, err := connectToDiscovery()
	if err != nil {
		return err
	}

	err = discovery.DeregisterService(name)
	if err != nil {
		log.Fatalf("failed to deregister with discovery service: %v", err)
		return err
	}

	return nil
}

func connectToDiscovery() (*Discovery, error) {
	discovery, err := NewDiscovery()
	if err != nil {
		log.Fatalf("failed to create discovery service: %v", err)
		return nil, err
	}

	return discovery, nil
}
