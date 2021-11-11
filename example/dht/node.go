package dht

import (
	"fmt"
	"hash/crc32"
	"sync"
)

// NewNode .
func NewNode(data interface{}) (result *Node) {
	crc32Value := crc32.ChecksumIEEE([]byte(fmt.Sprint(data)))
	result = &Node{data: data, crc32Value: crc32Value, property: make(map[string]interface{})}
	return
}

func newNodes(data []interface{}) (result []*Node) {
	replicateMap := map[interface{}]struct{}{}
	for _, key := range data {
		if _, ok := replicateMap[key]; ok {
			panic("Repeated nodes :" + fmt.Sprint(key))
		}
		replicateMap[key] = struct{}{}
	}

	for _, v := range data {
		crc32Value := crc32.ChecksumIEEE([]byte(fmt.Sprint(v)))
		result = append(result, &Node{data: v, crc32Value: crc32Value, property: make(map[string]interface{})})
	}
	return
}

// Node .
type Node struct {
	data       interface{}
	crc32Value uint32
	property   map[string]interface{}
	mutex      sync.RWMutex
}

// CRC32 .
func (node *Node) CRC32() uint32 {
	return node.crc32Value
}

// String .
func (node *Node) String() string {
	return fmt.Sprint(node.data)
}

// Value .
func (node *Node) Value() interface{} {
	return node.data
}

// SetProperty .
func (node *Node) SetProperty(key string, value interface{}) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.property[key] = value
}

// GetProperty .
func (node *Node) GetProperty(key string) (result interface{}, ok bool) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	result, ok = node.property[key]
	return
}
