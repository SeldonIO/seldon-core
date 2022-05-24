package util

import (
	"github.com/serialx/hashring"
)

type LoadBalancer interface {
	AddServer(serverName string)
	RemoveServer(serverName string)
	GetServersForKey(key string) []string
}

type RingLoadBalancer struct {
	ring              *hashring.HashRing
	nodes             map[string]bool
	replicationFactor int
}

func NewRingLoadBalancer(replicationFactor int) *RingLoadBalancer {
	return &RingLoadBalancer{
		ring:              hashring.New([]string{}),
		replicationFactor: replicationFactor,
		nodes:             make(map[string]bool),
	}
}

func (lb *RingLoadBalancer) AddServer(serverName string) {
	lb.ring = lb.ring.AddNode(serverName)
	lb.nodes[serverName] = true
}

func (lb *RingLoadBalancer) RemoveServer(serverName string) {
	lb.ring = lb.ring.RemoveNode(serverName)
	delete(lb.nodes, serverName)
}

func (lb *RingLoadBalancer) allKeys() []string {
	keys := make([]string, len(lb.nodes))
	i := 0
	for k := range lb.nodes {
		keys[i] = k
		i++
	}
	return keys
}

func (lb *RingLoadBalancer) GetServersForKey(key string) []string {
	if len(lb.nodes) < lb.replicationFactor {
		return lb.allKeys()
	} else {
		nodes, _ := lb.ring.GetNodes(key, lb.replicationFactor)
		return nodes
	}
}
