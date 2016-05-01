package rpc

import (
	"time"

	log "github.com/Sirupsen/logrus"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"net"
	"strconv"
)

var (
	usageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pi",
		Subsystem: "discovery",
		Name:      "usage_total",
		Help:      "Number of times endpoints has been invoked.",
	}, []string{"method"})
	responseTimeSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "pi",
		Subsystem: "discovery",
		Name:      "response_time",
		Help:      "Response time of endpoints.",
	}, []string{"method"})
)

func init() {
	prometheus.MustRegisterOrGet(usageCounter)
	prometheus.MustRegisterOrGet(responseTimeSummary)
}

// Node represents the address of an instance of a service
type Node struct {
	Address string
	Port    int
}

// Service represents a service
type Service struct {
	Name string
	Node Node
}

// Discovery is the main interface to all things discovery
type Discovery struct {
	client *consulapi.Client
}

// NewDiscovery returns a new discovery initialized
func NewDiscovery() (*Discovery, error) {
	consulAPIClient, err := newConsulAPIClient()
	if err != nil {
		return nil, err
	}

	return &Discovery{
		client: consulAPIClient,
	}, nil
}

// GetLocalServices is used to get all the services managed by the local agent
func (d *Discovery) GetLocalServices() ([]*Service, error) {
	log.Debug("-> discovery.GetLocalServices")
	start := time.Now()
	usageCounter.WithLabelValues("GetLocalServices").Inc()

	agentServices, err := d.client.Agent().Services()
	if err != nil {
		return nil, err
	}

	var services []*Service
	for _, agentService := range agentServices {
		var service Service
		service.Name = agentService.ID
		var node Node
		node.Address = agentService.Address
		node.Port = agentService.Port
		service.Node = node
		log.WithFields(log.Fields{
			"name":    service.Name,
			"address": service.Node.Address,
			"port":    service.Node.Port,
		}).Debug("GetLocalServices found service")

		services = append(services, &service)
	}

	elapsed := float64(time.Since(start)) / float64(time.Microsecond)
	responseTimeSummary.WithLabelValues("GetLocalServices").Observe(elapsed)

	log.WithField("elapsed", elapsed).Debug("<- discovery.GetLocalServices")
	return services, nil
}

// GetService is used to get all instances of a service
func (d *Discovery) GetService(name string) ([]*Node, error) {
	log.WithField("name", name).Debug("-> discovery.GetService")
	start := time.Now()
	usageCounter.WithLabelValues("GetService").Inc()

	serviceEntries, _, err := d.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, len(serviceEntries))
	for index, serviceEntry := range serviceEntries {
		var node Node
		node.Address = serviceEntry.Node.Address
		node.Port = serviceEntry.Service.Port
		log.WithFields(log.Fields{
			"name":    serviceEntry.Service.ID,
			"address": node.Address,
			"port":    node.Port,
		}).Debug("GetService found node")

		nodes[index] = &node
	}

	elapsed := float64(time.Since(start)) / float64(time.Microsecond)
	responseTimeSummary.WithLabelValues("GetService").Observe(elapsed)

	log.WithField("elapsed", elapsed).Debug("<- discovery.GetService")
	return nodes, nil
}

func (d *Discovery) GetRandomServiceAddress(name string) (*string, error) {
	log.WithField("name", name).Debug("-> discovery.GetRandomServiceNode")
	start := time.Now()
	usageCounter.WithLabelValues("GetRandomServiceNode").Inc()

	nodes, err := d.GetService(name)
	if err != nil {
		return nil, err
	}

	randomInt := rand.Intn(len(nodes))
	node := nodes[randomInt]
	log.WithFields(log.Fields{
		"total":        len(nodes),
		"index_chosen": randomInt,
		"address":      node.Address,
		"port": 	node.Port,
	}).Info("randomly chose node")

	address := net.JoinHostPort(node.Address, strconv.Itoa(node.Port))

	elapsed := float64(time.Since(start)) / float64(time.Microsecond)
	responseTimeSummary.WithLabelValues("GetRandomServiceNode").Observe(elapsed)

	log.WithField("elapsed", elapsed).Debug("<- discovery.GetRandomServiceNode")
	return &address, nil
}

// RegisterService is used to register a service with discovery service
func (d *Discovery) RegisterService(name string, port int) error {
	log.WithField("name", name).WithField("port", port).Debug("-> discovery.RegisterService")
	start := time.Now()
	usageCounter.WithLabelValues("RegisterService").Inc()

	var agentServiceRegistration consulapi.AgentServiceRegistration
	agentServiceRegistration.Name = name
	agentServiceRegistration.Port = port
	err := d.client.Agent().ServiceRegister(&agentServiceRegistration)
	if err != nil {
		return err
	}

	elapsed := float64(time.Since(start)) / float64(time.Microsecond)
	responseTimeSummary.WithLabelValues("RegisterService").Observe(elapsed)

	log.WithField("elapsed", elapsed).Debug("<- discovery.RegisterService")
	return nil
}

// DeregisterService is used to deregister a service with discovery service
func (d *Discovery) DeregisterService(name string) error {
	log.WithField("name", name).Debug("-> discovery.DeregisterService")
	start := time.Now()
	usageCounter.WithLabelValues("DeregisterService").Inc()

	err := d.client.Agent().ServiceDeregister(name)
	if err != nil {
		return err
	}

	elapsed := float64(time.Since(start)) / float64(time.Microsecond)
	responseTimeSummary.WithLabelValues("DeregisterService").Observe(elapsed)

	log.WithField("elapsed", elapsed).Debug("<- discovery.DeregisterService")
	return nil
}

func newConsulAPIClient() (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.HttpClient.Timeout = 2 * time.Second
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
