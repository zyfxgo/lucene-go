package index

import (
	"errors"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/packed"
	"math"
)

const (
	HAS_VALUE_MASK    = 1
	HAS_NO_VALUE_MASK = 0
	SHIFT             = 1
)

// DocValuesFieldUpdates
// Holds updates of a single DocValues field, for a set of documents within one segment.
// lucene.experimental
type DocValuesFieldUpdates interface {
	Field() string
	AddInt64(doc int, value int64) error
	AddBytes(doc int, value []byte) error

	// AddIterator
	// Adds the value for the given docID. This method prevents conditional calls to
	// DocValuesFieldUpdates.Iterator.longValue() or DocValuesFieldUpdates.Iterator.binaryValue()
	// since the implementation knows if it's a long value iterator or binary value
	AddIterator(doc int, value DocValuesFieldUpdatesIterator) error

	// Iterator
	// Returns an DocValuesFieldUpdates.Iterator over the updated documents and their values.
	Iterator() (DocValuesFieldUpdatesIterator, error)

	// Finish
	// Freezes internal data structures and sorts updates by docID for efficient iteration.
	Finish() error

	// Any
	// Returns true if this instance contains any updates.
	Any() bool

	Size() int

	// Reset
	// Adds an update that resets the documents value.
	// Params: doc – the doc to update
	Reset(doc int) error

	Swap(i, j int) error

	Grow(i int) error

	Resize(i int) error

	EnsureFinished() error
	GetFinished() bool
}

type DocValuesFieldUpdatesDefault struct {
	field        string
	_type        types.DocValuesType
	delGen       int64
	bitsPerValue int
	finished     bool
	maxDoc       int
	docs         *packed.PagedMutable
	size         int
}

func (d *DocValuesFieldUpdatesDefault) Field() string {
	return d.field
}

func (d *DocValuesFieldUpdatesDefault) Finish() error {
	if d.finished {
		return errors.New("already finished")
	}
	d.finished = true
	// shrink wrap
	if d.size < d.docs.Size() {
		d.Resize(d.size)
	}
	if d.size > 0 {
		// We need a stable sort but InPlaceMergeSorter performs lots of swaps
		// which hurts performance due to all the packed ints we are using.
		// Another option would be TimSorter, but it needs additional API (copy to
		// temp storage, compare with item in temp storage, etc.) so we instead
		// use quicksort and record ords of each update to guarantee stability.
		ords := packed.PackedIntsGetMutable(d.size, packed.PackedIntsBitsRequired(uint64(d.size-1)), packed.DEFAULT)
		for i := 0; i < d.size; i++ {
			ords.Set(i, uint64(i))
		}

	}

	return nil
}

// Any Returns true if this instance contains any updates.
func (d *DocValuesFieldUpdatesDefault) Any() bool {
	return d.size > 0
}

func (d *DocValuesFieldUpdatesDefault) Size() int {
	return d.size
}

func (d *DocValuesFieldUpdatesDefault) Swap(i, j int) error {
	tmpDoc := d.docs.Get(j)
	d.docs.Set(j, d.docs.Get(i))
	d.docs.Set(i, tmpDoc)
	return nil
}

func (d *DocValuesFieldUpdatesDefault) Grow(size int) error {
	d.docs = d.docs.Grow(size).(*packed.PagedMutable)
	return nil
}

func (d *DocValuesFieldUpdatesDefault) Resize(size int) error {
	d.docs = d.docs.Resize(size).(*packed.PagedMutable)
	return nil
}

func (d *DocValuesFieldUpdatesDefault) GetFinished() bool {
	return d.finished
}

func (b *BinaryDocValuesFieldUpdates) add(doc int) (int, error) {
	return b.addInternal(doc, HAS_VALUE_MASK)
}

func (b *BinaryDocValuesFieldUpdates) addInternal(doc int, hasValueMask int64) (int, error) {
	if b.finished {
		return 0, errors.New("already finished")
	}

	if doc >= b.maxDoc {
		return 0, errors.New("doc too big")
	}

	// TODO: if the Sorter interface changes to take long indexes, we can remove that limitation
	if b.size == math.MaxInt32 {
		return 0, errors.New("cannot support more than Integer.MAX_VALUE doc/value entries")
	}

	// grow the structures to have room for more elements
	if b.docs.Size() == b.size {
		if err := b.Grow(b.size + 1); err != nil {
			return 0, err
		}
	}

	value := (int64(doc) << SHIFT) | hasValueMask
	b.docs.Set(b.size, uint64(value))
	b.size++
	return b.size - 1, nil
}

// DocValuesFieldUpdatesIterator
// An iterator over documents and their updated values. Only documents with updates are returned
// by this iterator, and the documents are returned in increasing order.
type DocValuesFieldUpdatesIterator interface {
	DocValuesIterator

	// LongValue Returns a long value for the current document if this iterator is a long iterator.
	LongValue() (int64, error)

	// BinaryValue Returns a binary value for the current document if this iterator is a binary value iterator.
	BinaryValue() ([]byte, error)

	// DelGen Returns delGen for this packet.
	DelGen() int64

	// HasValue Returns true if this doc has a value
	HasValue() bool
}

type DVFUIterator struct {
}

func (*DVFUIterator) AdvanceExact(target int) (bool, error) {
	return false, errors.New("unsupported operation exception")
}

func (*DVFUIterator) Advance(target int) (int, error) {
	return 0, errors.New("unsupported operation exception")
}

func (*DVFUIterator) Cost() int64 {
	return 0
}

func AsBinaryDocValues(iterator DocValuesFieldUpdatesIterator) BinaryDocValues {
	return &BinaryDocValuesDefault{
		FnDocID:        iterator.DocID,
		FnNextDoc:      iterator.NextDoc,
		FnAdvance:      iterator.Advance,
		FnSlowAdvance:  iterator.SlowAdvance,
		FnCost:         iterator.Cost,
		FnAdvanceExact: iterator.AdvanceExact,
		FnBinaryValue:  iterator.BinaryValue,
	}
}

func AsNumericDocValues(iterator DocValuesFieldUpdatesIterator) NumericDocValues {
	return &NumericDocValuesDefault{
		FnDocID:        iterator.DocID,
		FnNextDoc:      iterator.NextDoc,
		FnAdvance:      iterator.Advance,
		FnSlowAdvance:  iterator.SlowAdvance,
		FnCost:         iterator.Cost,
		FnAdvanceExact: iterator.AdvanceExact,
		FnLongValue:    iterator.LongValue,
	}
}

type SingleValueDocValuesFieldUpdates struct {
}