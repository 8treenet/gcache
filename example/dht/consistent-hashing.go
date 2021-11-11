package dht

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

// New .
func New() DHT {
	result := &dht{}
	return result
}

// DHT Create Consistent Hashing Object
type DHT interface {
	Range(begin, end int) *ConsistentHashing
	List(data ...interface{}) *ConsistentHashing
}

type dht struct{}

// Range The object that creates the range node
func (hash *dht) Range(begin, end int) *ConsistentHashing {
	if end < begin {
		panic("The range must be in positive order")
	}

	data := []interface{}{}
	for i := begin; i <= end; i++ {
		data = append(data, i)
	}

	nodes := newNodes(data)
	result := &ConsistentHashing{nodes: nodes}
	result.sort()
	return result
}

// List The object that created the list node
func (hash *dht) List(data ...interface{}) *ConsistentHashing {
	if len(data) == 0 {
		panic("The data cannot be empty")
	}

	nodes := newNodes(data)
	result := &ConsistentHashing{nodes: nodes}
	result.sort()
	return result
}

// ConsistentHashing .
type ConsistentHashing struct {
	nodes []*Node
	mutex sync.RWMutex
}

// Search .
func (hash *ConsistentHashing) Search(data interface{}) *Node {
	hash.mutex.RLock()
	defer hash.mutex.RUnlock()

	if len(hash.nodes) == 0 {
		return nil
	}
	crc32Value := crc32.ChecksumIEEE([]byte(fmt.Sprint(data)))
	index := sort.Search(len(hash.nodes), func(i int) bool {
		return hash.nodes[i].crc32Value <= crc32Value
	})

	if index == len(hash.nodes) {
		return hash.nodes[0]
	}
	return hash.nodes[index]
}

// GetNodes .
func (hash *ConsistentHashing) GetNodes() (result []*Node) {
	hash.mutex.RLock()
	defer hash.mutex.RUnlock()

	result = append(result, hash.nodes...)
	return
}

// AddNode .
func (hash *ConsistentHashing) AddNode(node *Node) error {
	hash.mutex.Lock()
	defer hash.mutex.Unlock()

	for _, v := range hash.nodes {
		if v.crc32Value == node.crc32Value {
			return errors.New("Repeated nodes :" + fmt.Sprint(node.data))
		}
	}

	hash.nodes = append(hash.nodes, node)
	hash.sort()
	return nil
}

// RemoveNode .
func (hash *ConsistentHashing) RemoveNode(node *Node) {
	hash.mutex.Lock()
	defer hash.mutex.Unlock()

	newNodes := []*Node{}
	found := false
	for _, v := range hash.nodes {
		if v.crc32Value == node.crc32Value {
			found = true
			continue
		}
		newNodes = append(newNodes, v)
	}

	if !found {
		return
	}
	hash.nodes = newNodes
	hash.sort()
}

// FindNode .
func (hash *ConsistentHashing) FindNode(data interface{}) (result *Node) {
	hash.mutex.RLock()
	defer hash.mutex.RUnlock()

	for _, v := range hash.nodes {
		if v.data == data {
			return v
		}
	}
	return nil
}

func (hash *ConsistentHashing) sort() {
	sort.Slice(hash.nodes, func(i, j int) bool {
		return hash.nodes[i].crc32Value > hash.nodes[j].crc32Value
	})
}
