package fst

import "io"

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	io.Reader

	// GetPosition Get current read position.
	GetPosition() int64

	// SetPosition Set current read position.
	SetPosition(pos int64)

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}