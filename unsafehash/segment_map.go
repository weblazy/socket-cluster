package unsafehash

import (
	"sort"
)

type HashRing []int64

func (c HashRing) Len() int {
	return len(c)
}

func (c HashRing) Less(i, j int) bool {
	return c[i] < c[j]
}

func (c HashRing) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type Node struct {
	Id       string
	Position int64
	Extra    interface{}
}

func NewNode(id string, position int64, extra interface{}) *Node {
	return &Node{
		Id:       id,
		Position: position,
		Extra:    extra,
	}
}

type SegmentMap struct {
	Nodes     map[int64]*Node
	Resources map[string]bool
	ring      HashRing
}

func NewSegmentMap() *SegmentMap {
	return &SegmentMap{
		Nodes:     make(map[int64]*Node),
		Resources: make(map[string]bool),
		ring:      HashRing{},
	}
}

func (c *SegmentMap) Add(node *Node) bool {
	if _, ok := c.Resources[node.Id]; ok {
		return false
	}
	c.Nodes[node.Position] = node
	c.Resources[node.Id] = true
	c.sortHashRing()
	return true
}

func (c *SegmentMap) sortHashRing() {
	c.ring = HashRing{}
	for k := range c.Nodes {
		c.ring = append(c.ring, k)
	}
	sort.Sort(c.ring)
}

func (c *SegmentMap) Get(key int64) *Node {
	i := c.search(key)

	return c.Nodes[c.ring[i]]
}

func (c *SegmentMap) search(hash int64) int {
	n := len(c.ring)
	i := sort.Search(n, func(i int) bool { return c.ring[i] >= hash })
	if i < n {
		return i
	} else {
		return 0
	}
}

func (c *SegmentMap) Remove(node *Node) {
	if _, ok := c.Resources[node.Id]; !ok {
		return
	}
	delete(c.Resources, node.Id)
	delete(c.Nodes, node.Position)
	c.sortHashRing()
}
