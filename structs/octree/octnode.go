package octree

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// Contains data necessary to build the octree
type OctNode struct {
	Parent              *OctNode
	BoundingBox         *BoundingBox
	Children            [8]*OctNode
	Items               []*OctElement
	Depth               uint8
	GlobalChildrenCount int64
	LocalChildrenCount  int32
	Opts                *TilerOptions
	IsLeaf              bool
	sync.RWMutex
}

// Instantiates a new OctNode properly initializing the data
func NewOctNode(boundingBox *BoundingBox, opts *TilerOptions, depth uint8, parent *OctNode) *OctNode {
	octNode := OctNode{
		Parent:              parent,
		BoundingBox:         boundingBox,
		Depth:               depth,
		Opts:                opts,
		GlobalChildrenCount: 0,
		IsLeaf:              true,
	}
	return &octNode
}

// Adds an OctElement to the OctNode eventually propagating it to the OctNode relevant children
func (octNode *OctNode) AddOctElement(element *OctElement) {
	if atomic.LoadInt32(&octNode.LocalChildrenCount) < octNode.Opts.MaxNumPointsPerNode {
		octNode.Lock()
		octNode.Items = append(octNode.Items, element)
		atomic.AddInt32(&octNode.LocalChildrenCount, 1)
		octNode.Unlock()
	} else {
		octNode.getSubOctNodeContainingElement(element).AddOctElement(element)
	}
	atomic.AddInt64(&octNode.GlobalChildrenCount, 1)
}

// Gets the children OctNode deemed to contain the given OctElement according to its coordinates
func (octNode *OctNode) getSubOctNodeContainingElement(element *OctElement) *OctNode {
	octant := octNode.BoundingBox.getOctantFromElement(element)

	// Acquire read lock on node
	octNode.RLock()
	child := octNode.Children[octant]
	octNode.RUnlock()
	if child != nil {
		return child
	}

	// Child not found. Create it.
	// First acquire Write lock and defer lock release
	octNode.Lock()
	defer octNode.Unlock()
	if octNode.Children[octant] == nil {
		newDepth := octNode.Depth + 1
		octNode.Children[octant] = NewOctNode(octNode.BoundingBox.getOctantBoundingBox(&octant), octNode.Opts, newDepth, octNode)
		octNode.IsLeaf = false
	}
	return octNode.Children[octant]
}

// Prints the summary of the node contents in the console
func (octNode *OctNode) PrintStructure() {
	fmt.Println(strings.Repeat(" ", int(octNode.Depth)-1)+"-", "element no:", octNode.LocalChildrenCount, "leaf:", octNode.IsLeaf)
	for _, e := range octNode.Children {
		if e != nil {
			e.PrintStructure()
		}
	}
}