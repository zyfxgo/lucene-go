package bkd

import "github.com/geange/lucene-go/core/store"

type BKDReader struct {

	// Packed array of byte[] holding all split values in the full binary tree:
	leafNodeOffset int
	config         *BKDConfig
	numLeaves      int
	in             store.IndexInput
	minPackedValue []byte
	maxPackedValue []byte
	pointCount     int64
	docCount       int
	version        int
	minLeafBlockFP int64

	packedIndex store.IndexInput
}
