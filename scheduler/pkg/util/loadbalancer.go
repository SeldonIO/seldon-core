/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"sync"

	"github.com/serialx/hashring"
)

type LoadBalancer interface {
	AddServer(serverName string)
	RemoveServer(serverName string)
	UpdatePartitions(numPartitions int)
	GetServersForKey(key string) []string
}

type RingLoadBalancer struct {
	ring              *hashring.HashRing
	mu                sync.RWMutex
	nodes             map[string]bool
	replicationFactor int
	numPartitions     int
}

func NewRingLoadBalancer(numPartitions int) *RingLoadBalancer {
	return &RingLoadBalancer{
		ring:          hashring.New([]string{}),
		nodes:         make(map[string]bool),
		numPartitions: numPartitions,
	}
}

func (lb *RingLoadBalancer) UpdatePartitions(numPartitions int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.numPartitions = numPartitions
	lb.replicationFactor = min(len(lb.nodes), lb.numPartitions)
}

func (lb *RingLoadBalancer) AddServer(serverName string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.ring = lb.ring.AddNode(serverName)
	lb.nodes[serverName] = true
	lb.replicationFactor = min(len(lb.nodes), lb.numPartitions)
}

func (lb *RingLoadBalancer) RemoveServer(serverName string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.ring = lb.ring.RemoveNode(serverName)
	delete(lb.nodes, serverName)
	lb.replicationFactor = min(len(lb.nodes), lb.numPartitions)
}

func (lb *RingLoadBalancer) GetServersForKey(key string) []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	nodes, _ := lb.ring.GetNodes(key, lb.replicationFactor)
	return nodes
}
