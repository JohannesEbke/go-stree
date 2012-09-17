// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Copyright 2012 Johannes Ebke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package implements a segment tree and serial algorithm to query intervals
package stree

import (
	"fmt"
	"sort"
)

// Main interface to access tree
type Tree interface {
	// Push new interval to stack
	Push(from, to int)
	// Push array of intervals to stack
	PushArray(from, to []int)
	// Clear the interval stack
	Clear()
	// Build segment tree out of interval stack
	BuildTree()
	// Print tree recursively to stdout
	Print()
	// Transform tree to array
	Tree2Array() []SegmentOverlap
	// Query interval
	Query(from, to int) []Interval
	// Query interval array
	QueryArray(from, to []int) []Interval
}

type stree struct {
	// Node list
	nodes []node
	// Interval stack
	base []Interval
	// Min value of all intervals
	min int
	// Max value of all intervals
	max int
        // Temporary map to hold node overlaps
	temp_overlap map[int32][]int32
        // Permanent overlap store
        overlaps []int32
}

type node struct {
	// left and right children index. 32 bit offsets are enough for 128 GB stree.
	left, right int32
	// A segment is a interval represented by the node
	segment Segment
	// Offset into overlaps struct in stree
	overlap_index int32
}

// Accessor helper functions to look up child node indices
func (t *stree) rightChild(node int32) int32 {
	return t.nodes[node].right
}

func (t *stree) leftChild(node int32) int32 {
	return t.nodes[node].left
}

type Interval struct {
	Id int // unique
	Segment
}

type Segment struct {
	From int
	To   int
}

// Represents overlapping intervals of a segment
type SegmentOverlap struct {
	Segment  Segment
	Interval []Interval
}

// Node receiver for tree traversal
type NodeReceive func(*stree, int32)

const (
	// Relations of two intervals
	SUBSET = iota
	DISJOINT
	INTERSECT_OR_SUPERSET
)

// Traverse tree recursively call enter when entering node, resp. leave
func (t *stree) traverse(node int32, enter, leave NodeReceive) {
       if !t.HasNode(node) {
               return
       }
       if enter != nil {
               enter(t, node)
       }
       t.traverse(t.rightChild(node), enter, leave) // right
       t.traverse(t.leftChild(node), enter, leave)  // left
       if leave != nil {
               leave(t, node)
       }
}

func (t *stree) Print() {
       t.traverse(0, func(t *stree, node int32) {
               segment, overlap := t.GetNode(node)
               fmt.Printf("\nSegment: (%d,%d)", segment.From, segment.To)
               for intrvl := range overlap {
                       fmt.Printf("\nInterval %d: (%d,%d)", t.base[intrvl].Id, t.base[intrvl].Segment.From, t.base[intrvl].Segment.To)
               }
       }, nil)
}

// Tree2Array transforms tree to array
func (t *stree) Tree2Array() []SegmentOverlap {
       array := make([]SegmentOverlap, 0, 50)
       t.traverse(0, func(t *stree, node int32) {
               segment, _ := t.GetNode(node)
               seg := SegmentOverlap{Segment: *segment, Interval: t.Overlap(node)}
               array = append(array, seg)
       }, nil)
       return array
}

// Overlap transforms []*Interval to []Interval
func (t *stree) Overlap(node_index int32) []Interval {
       _, overlap := t.GetNode(node_index)
       if overlap == nil {
               return nil
       }
       interval := make([]Interval, len(overlap))
       for i, pintrvl := range overlap {
               interval[i] = t.base[pintrvl]
       }
       return interval
}

// NewTree returns a Tree interface with underlying segment tree implementation
func NewTree() Tree {
	t := new(stree)
	t.Clear()
	return t
}

// Push new interval to stack
func (t *stree) Push(from, to int) {
	t.base = append(t.base, Interval{len(t.base), Segment{from, to}})
}

// Push array of intervals to stack
func (t *stree) PushArray(from, to []int) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Clear the interval stack
func (t *stree) Clear() {
	t.nodes = make([]node, 0)
	t.base = make([]Interval, 0)
	t.min = 0
	t.max = 0
        t.temp_overlap = make(map[int32][]int32)
}

// Build segment tree out of interval stack
func (t *stree) BuildTree() {
	if len(t.base) == 0 {
		panic("No intervals in stack to build tree. Push intervals first")
	}
	var endpoint []int
	endpoint, t.min, t.max = Endpoints(t.base)
	// Create tree nodes from interval endpoints
	t.insertNodes(endpoint)
	for i := range t.base {
		t.insertInterval(0, int32(i))
	}
	for node := range t.nodes {
            t.nodes[int32(node)].overlap_index = int32(len(t.overlaps))
            overlap := t.temp_overlap[int32(node)]
            for i := range overlap {
                t.overlaps = append(t.overlaps, overlap[i])
            }
        }
        t.temp_overlap = nil
}

// Add a node to the tree
func (t *stree) AddNode(from, to int) int32 {
	node_id := int32(len(t.nodes))
	t.nodes = append(t.nodes, node{-1, -1, Segment{from, to}, 0})
	return node_id
}

// Ask if a node exists
func (t *stree) HasNode(node int32) bool {
	return node >= 0 && node < int32(len(t.nodes))
}

// Get a node from the tree
func (t *stree) GetNode(node int32) (*Segment, []int32) {
	if !t.HasNode(node) {
		return nil, nil
	}
        o_begin := int32(t.nodes[node].overlap_index)
        o_end := int32(len(t.overlaps))
        if node != int32(len(t.nodes) - 1) {
            o_end = t.nodes[node+1].overlap_index
        }
        if (o_begin == o_end) {
            return &t.nodes[node].segment, nil
        }
        return &t.nodes[node].segment, t.overlaps[o_begin:o_end]
}

func (t *stree) GetIntervalSegment(interval int32) *Segment {
	return &t.base[interval].Segment
}

// Endpoints returns a slice with all endpoints (sorted, unique)
func Endpoints(base []Interval) (result []int, min, max int) {
	baseLen := len(base)
	endpoints := make([]int, baseLen*2)
	for i, interval := range base {
		endpoints[i] = interval.Segment.From
		endpoints[i+baseLen] = interval.Segment.To
	}
	result = Dedup(endpoints)
	min = result[0]
	max = result[len(result)-1]
	return
}

// Dedup removes duplicates from a given slice
func Dedup(sl []int) []int {
	sort.Sort(sort.IntSlice(sl))
	unique := make([]int, 0, len(sl))
	prev := sl[0] + 1
	for _, val := range sl {
		if val != prev {
			unique = append(unique, val)
			prev = val
		}
	}
	return unique
}

// insertNodes builds tree structure from given endpoints
func (t *stree) insertNodes(endpoint []int) int32 {
	if len(endpoint) == 2 {
		node := t.AddNode(endpoint[0], endpoint[1])
		if endpoint[1] != t.max {
			t.nodes[node].left = t.AddNode(endpoint[0], endpoint[1])
			t.nodes[node].right = t.AddNode(endpoint[1], endpoint[1])
		}
		return node
	} else {
		node := t.AddNode(endpoint[0], endpoint[len(endpoint)-1])
		center := len(endpoint) / 2
		t.nodes[node].left = t.insertNodes(endpoint[:center+1]) // left
		t.nodes[node].right = t.insertNodes(endpoint[center:])  // right
		return node
	}
	return -1 // can not happen
}

// CompareTo compares two Segments and returns: DISJOINT, SUBSET or INTERSECT_OR_SUPERSET
func (s *Segment) CompareTo(other *Segment) int {
	if other.From > s.To || other.To < s.From {
		return DISJOINT
	}
	if other.From <= s.From && other.To >= s.To {
		return SUBSET
	}
	return INTERSECT_OR_SUPERSET
}

// Disjoint returns true if Segment does not overlap with interval
func (s *Segment) Disjoint(from, to int) bool {
	if from > s.To || to < s.From {
		return true
	}
	return false
}

// Inserts interval into given tree structure
func (t *stree) insertInterval(node int32, intrvl int32) {
	segment, _ := t.GetNode(node)
	switch segment.CompareTo(t.GetIntervalSegment(intrvl)) {
	case SUBSET:
		// interval of node is a subset of the specified interval or equal
		if t.temp_overlap[node] == nil {
			t.temp_overlap[node] = make([]int32, 1)
			t.temp_overlap[node][0] = intrvl
		} else {
			t.temp_overlap[node] = append(t.temp_overlap[node], intrvl)
		}
	case INTERSECT_OR_SUPERSET:
		// interval of node is a superset, have to look in both children
		if t.HasNode(t.leftChild(node)) { // left
			t.insertInterval(t.leftChild(node), intrvl)
		}
		if t.HasNode(t.rightChild(node)) {
			t.insertInterval(t.rightChild(node), intrvl) // right
		}
	case DISJOINT:
		// nothing to do
	}
}

// Query interval
func (t *stree) Query(from, to int) []Interval {
	result := make(map[int32]bool)
	t.querySingle(0, from, to, &result)
	// transform map to slice
	sl := make([]Interval, 0, len(result))
	for intrvl, _ := range result {
		sl = append(sl, t.base[intrvl])
	}
	return sl
}

// querySingle traverse tree in search of overlaps
func (t *stree) querySingle(node int32, from, to int, result *map[int32]bool) {
	segment, overlap := t.GetNode(node)
	if !segment.Disjoint(from, to) {
		for _, pintrvl := range overlap {
			(*result)[pintrvl] = true
		}
		if t.HasNode(t.rightChild(node)) {
			t.querySingle(t.rightChild(node), from, to, result)
		}
		if t.HasNode(t.leftChild(node)) {
			t.querySingle(t.leftChild(node), from, to, result)
		}
	}
}

// Query interval array
func (t *stree) QueryArray(from, to []int) []Interval {
	result := make(map[int32]bool)
	t.queryMulti(0, from, to, &result)
	sl := make([]Interval, 0, len(result))
	for intrvl, _ := range result {
		sl = append(sl, t.base[intrvl])
	}
	return sl
}

// queryMulti traverse tree in search of overlaps with multiple intervals
func (t *stree) queryMulti(node int32, from, to []int, result *map[int32]bool) {
	hitsFrom := make([]int, 0, 2)
	hitsTo := make([]int, 0, 2)
	segment, overlap := t.GetNode(node)
	for i, fromvalue := range from {
		if !segment.Disjoint(fromvalue, to[i]) {
			for _, pintrvl := range overlap {
				(*result)[pintrvl] = true
			}
			hitsFrom = append(hitsFrom, fromvalue)
			hitsTo = append(hitsTo, to[i])
		}
	}
	// search in children only with overlapping intervals of parent
	if len(hitsFrom) != 0 {
		if t.HasNode(t.rightChild(node)) {
			t.queryMulti(t.rightChild(node), hitsFrom, hitsTo, result)
		}
		if t.HasNode(t.leftChild(node)) {
			t.queryMulti(t.leftChild(node), hitsFrom, hitsTo, result)
		}
	}
}
