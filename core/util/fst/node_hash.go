package fst

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// NodeHash Used to dedup states (lookup already-frozen states)
type NodeHash[T any] struct {
	table map[int]int64
	//count      int64
	//mask       int64
	fst        *Fst[T]
	scratchArc *Arc[T]
	in         BytesReader
}

func NewNodeHash[T any](fst *Fst[T], in BytesReader) *NodeHash[T] {
	return &NodeHash[T]{
		table: make(map[int]int64),
		//mask:       15,
		fst:        fst,
		scratchArc: &Arc[T]{},
		in:         in,
	}
}

const (
	PRIME = int64(32)
)

func (n *NodeHash[T]) nodesEqual(node *UnCompiledNode[T], address int64) bool {
	_, err := n.fst.ReadFirstRealTargetArc(address, n.scratchArc, n.in)
	if err != nil {
		return false
	}

	// Fail fast for a node with fixed length arcs.
	if n.scratchArc.BytesPerArc() != 0 {
		if n.scratchArc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
			if node.NumArcs() != n.scratchArc.NumArcs() {
				return false
			}
		} else {
			{
				if n.scratchArc.NodeFlags() != ARCS_FOR_DIRECT_ADDRESSING {
					panic("")
				}
			}

			if int64(node.Arcs[len(node.Arcs)-1].Label-node.Arcs[0].Label+1) != n.scratchArc.NumArcs() {
				return false
			} else if v, err := CountBits(n.scratchArc, n.in); err == nil && v != node.NumArcs() {
				return false
			}
		}
	}

	for i := range node.Arcs {
		arc := node.Arcs[i]
		if arc.Label != n.scratchArc.Label() ||
			!(reflect.DeepEqual(arc.Output, n.scratchArc.output)) ||
			arc.Target.(*CompiledNode).node != n.scratchArc.Target() ||
			!reflect.DeepEqual(arc.NextFinalOutput, n.scratchArc.NextFinalOutput()) ||
			arc.IsFinal != n.scratchArc.IsFinal() {
			return false
		}

		if n.scratchArc.IsLast() {
			if i == int(node.NumArcs()-1) {
				return true
			}
			return false
		}

		_, err := n.fst.ReadNextRealArc(n.scratchArc, n.in)
		if err != nil {
			return false
		}
	}

	return false
}

// hashNode code for an unfrozen node.  This must be identical
// to the frozen case (below)!!
func (n *NodeHash[T]) hashUnfrozenNode(node *UnCompiledNode[T]) (int64, error) {
	h := int64(0)
	// TODO: maybe if number of arcs is high we can safely subsample?

	for i := range node.Arcs {
		arc := node.Arcs[i]
		h = PRIME*h + int64(arc.Label)

		target, ok := arc.Target.(*CompiledNode)
		if !ok {
			return 0, errors.New("target is not CompiledNode")
		}

		nodeValue := target.node

		h = PRIME*h + (nodeValue ^ (nodeValue >> 32))
		h = PRIME*h + hashObj(arc.Output)
		h = PRIME*h + hashObj(arc.NextFinalOutput)
		if arc.IsFinal {
			h += 17
		}
	}
	return h & math.MaxInt64, nil
}

func (n *NodeHash[T]) hashFrozenNode(node int64) (int64, error) {
	h := int64(0)
	var err error
	_, err = n.fst.ReadFirstRealTargetArc(node, n.scratchArc, n.in)
	if err != nil {
		return 0, err
	}

	for {
		h = PRIME*h + int64(n.scratchArc.Label())
		h = PRIME*h + (n.scratchArc.Target() ^ (n.scratchArc.Target() >> 32))
		h = PRIME*h + hashObj(n.scratchArc.Output())
		h = PRIME*h + hashObj(n.scratchArc.NextFinalOutput())

		if n.scratchArc.IsFinal() {
			h += 17
		}

		if n.scratchArc.IsLast() {
			break
		}
		_, err := n.fst.ReadNextRealArc(n.scratchArc, n.in)
		if err != nil {
			return 0, err
		}
	}

	return h & math.MaxInt64, nil
}

func (n *NodeHash[T]) Add(builder *Builder[T], nodeIn *UnCompiledNode[T]) (int64, error) {
	h, err := n.hashUnfrozenNode(nodeIn)
	if err != nil {
		return 0, err
	}
	//pos := h & n.mask
	pos := int(h)

	for {
		v, ok := n.table[pos]
		if !ok {
			// freeze & add
			node, err := n.fst.AddNode(builder, nodeIn)
			if err != nil {
				return 0, err
			}

			{
				frozenNode, err := n.hashFrozenNode(node)
				if err != nil {
					return 0, err
				}

				if frozenNode != h {
					return 0, fmt.Errorf("frozenHash=%d vs h=%d", frozenNode, h)
				}
			}

			//n.count++
			n.table[pos] = node
			//// Rehash at 2/3 occupancy:
			//if n.count > int64(2*n.table.Size()/3) {
			//	err := n.rehash()
			//	if err != nil {
			//		return 0, err
			//	}
			//}
			return node, nil
		}

		if n.nodesEqual(nodeIn, v) {
			return v, nil
		}
		pos++
	}
}

// called only by rehash
func (n *NodeHash[T]) addNew(address int64) error {
	v, err := n.hashFrozenNode(address)
	if err != nil {
		return err
	}
	//pos := v & n.mask
	pos := int(v)
	//c := int64(0)
	for {
		if _, ok := n.table[pos]; ok {
			n.table[pos] = address
			break
		}

		// quadratic probe
		//pos = (pos + (c + 1)) & n.mask
		//c++
		pos++
	}
	return nil
}

func hashObj(obj interface{}) int64 {
	// TODO: != noOutput
	if obj != nil {
		code, ok := obj.(HashCode)
		if ok {
			return code.Hash()
		}

		switch obj.(type) {
		case []byte:
			h := int64(0)
			for _, b := range obj.([]byte) {
				h = PRIME*h + int64(b)
			}
			return h
		case int64:
			value := obj.(int64)
			return value ^ (value >> 32)
		}

	}
	return 0
}

type HashCode interface {
	Hash() int64
}
