package fst

import (
	"github.com/geange/lucene-go/core/store"
)

type Box[T any] struct {
	Value T
}

// Outputs Represents the output for an Fst, providing the basic algebra required for building and traversing the Fst.
// Note that any operation that returns noOutput must return the same singleton object from getNoOutput.
//
// lucene.experimental
type Outputs[T any] interface {

	// Common Eg common("foobar", "food") -> "foo"
	Common(output1, output2 T) (T, error)

	// Subtract Eg subtract("foobar", "foo") -> "bar"
	Subtract(output1, inc T) (T, error)

	// Add Eg add("foo", "bar") -> "foobar"
	Add(prefix, output T) (T, error)

	// Write Encode an output value into a DataOutput.
	Write(output T, out store.DataOutput) error

	// WriteFinalOutput Encode an final node output value into a DataOutput. By default this just calls write(Object, DataOutput).
	WriteFinalOutput(output T, out store.DataOutput) error

	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(in store.DataInput) (T, error)

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(in store.DataInput) error

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(in store.DataInput) (T, error)

	// SkipFinalOutput Skip the output previously written with writeFinalOutput;
	// defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(in store.DataInput) error

	IsNoOutput(v T) bool

	GetNoOutput() T

	Merge(first, second T) (T, error)
}
