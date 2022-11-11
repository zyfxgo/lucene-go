package fst

import "io"

// Outputs Represents the outputs for an FST, providing the basic algebra required for building and traversing the FST.
// Note that any operation that returns NO_OUTPUT must return the same singleton object from getNoOutput.
//
// lucene.experimental
type Outputs interface {

	// Common Eg common("foobar", "food") -> "foo"
	Common(output1, output2 any) (any, error)

	// Subtract Eg subtract("foobar", "foo") -> "bar"
	Subtract(output1, inc any) (any, error)

	// Add Eg add("foo", "bar") -> "foobar"
	Add(prefix, output any) (any, error)

	// Write Encode an output value into a DataOutput.
	Write(output any, out io.Writer) error

	// Encode an final node output value into a DataOutput. By default this just calls write(Object, DataOutput).
	writeFinalOutput(output any, out io.Writer) error

	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(in io.Reader) (any, error)

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(in io.Reader) error

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(in io.Reader) (any, error)

	// SkipFinalOutput Skip the output previously written with writeFinalOutput;
	// defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(in io.Reader) error

	// GetNoOutput NOTE: this output is compared with == so you must ensure that all methods return
	// the single object if it's really no output
	GetNoOutput() any

	Merge(first, second any) (any, error)
}