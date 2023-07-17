package fst

import (
	"bytes"
	"errors"

	"github.com/geange/gods-generic/lists/arraylist"
	"github.com/geange/lucene-go/core/store"
)

// TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which Fst needs

type ByteStore struct {
	*store.WriterX

	blocks *arraylist.List[[]byte]

	blockSize int64
	blockBits int64
	blockMask int64

	current []byte

	nextWrite int64
}

func NewByteStore(blockBits int) *ByteStore {
	blockSize := int64(1 << blockBits)
	byteStore := &ByteStore{
		blocks:    arraylist.NewWith[[]byte](bytes.Compare),
		blockSize: blockSize,
		blockBits: int64(blockBits),
		blockMask: blockSize - 1,
		current:   make([]byte, blockSize),
		nextWrite: 0,
	}
	byteStore.blocks.Add(byteStore.current)
	byteStore.WriterX = store.NewWriterX(byteStore)
	return byteStore
}

var ErrItemNotFound = errors.New("item not found")

func NewBytesStoreV1(in store.DataInput, numBytes, maxBlockSize int64) (*ByteStore, error) {
	blockSize := int64(2)
	blockBits := int64(1)
	for blockSize < numBytes && blockSize < maxBlockSize {
		blockSize *= 2
		blockBits++
	}

	bs := &ByteStore{
		blocks: arraylist.NewWith[[]byte](bytes.Compare),
	}

	bs.blockBits = blockBits
	bs.blockSize = blockSize
	bs.blockMask = blockSize - 1
	left := numBytes
	for left > 0 {
		chunk := min(blockSize, left)
		block := make([]byte, chunk)
		_, err := in.Read(block)
		if err != nil {
			return nil, err
		}

		bs.blocks.Add(block)
		left -= chunk
	}

	// So .getPosition still works
	lastItem, ok := bs.blocks.Get(bs.blocks.Size() - 1)
	if !ok {
		return nil, errors.New("value not found")
	}
	bs.nextWrite = int64(len(lastItem))
	return bs, nil
}

// WriteByteAt Absolute write byte; you must ensure dest is < max position written so far.
func (r *ByteStore) WriteByteAt(dest int64, b byte) error {
	blockIndex := int64(dest >> r.blockBits)
	block, ok := r.blocks.Get(int(blockIndex))
	if !ok {
		return ErrItemNotFound
	}
	block[dest&r.blockMask] = b
	return nil
}

func (r *ByteStore) WriteByte(b byte) error {
	if r.nextWrite == r.blockSize || len(r.current) == 0 {
		r.current = make([]byte, r.blockSize)
		r.blocks.Add(r.current)
	}
	r.current[r.nextWrite] = b
	r.nextWrite++
	return nil
}

func (r *ByteStore) Write(bs []byte) (int, error) {
	if r.current == nil {
		r.current = make([]byte, r.blockSize)
		r.blocks.Add(r.current)
		r.nextWrite = 0
	}

	offset := int64(0)
	count := len(bs)
	size := int64(count)

	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			copy(r.current[r.nextWrite:], bs[offset:offset+size])
			r.nextWrite += size
			break
		} else {
			if chunk > 0 {
				copy(r.current[r.nextWrite:], bs[offset:offset+chunk])
				offset += chunk
				size -= chunk
			}
			r.current = make([]byte, r.blockSize)
			r.blocks.Add(r.current)
			r.nextWrite = 0
		}
	}
	return count, nil
}

func (r *ByteStore) GetBlockBits() int64 {
	return r.blockBits
}

// WriteBytesAt Absolute writeBytes without changing the current position.
// Note: this cannot "grow" the bytes, so you must only call it on already written parts.
func (r *ByteStore) WriteBytesAt(dest int64, bs []byte) error {
	size := int64(len(bs))

	end := dest + int64(len(bs))
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}

	block, ok := r.blocks.Get(int(blockIndex))
	if !ok {
		return ErrItemNotFound
	}

	for size > 0 {
		if size <= downTo {
			copy(block[downTo-size:downTo], bs)
			break
		}

		size -= downTo
		copy(block[0:downTo], bs[size:])
		blockIndex--

		block, ok = r.blocks.Get(int(blockIndex))
		if !ok {
			return ErrItemNotFound
		}
		downTo = r.blockSize
	}
	return nil
}

// MoveBytes Absolute copy bytes self to self, without changing the position.
// Note: this cannot "grow" the bytes, so must only call it on already written parts.
func (r *ByteStore) MoveBytes(src, dest, size int64) error {
	if src >= dest {
		return errors.New("src >= dest")
	}

	end := src + size
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}

	block, ok := r.blocks.Get(int(blockIndex))
	if !ok {
		return ErrItemNotFound
	}

	for size > 0 {
		if size <= downTo {
			err := r.WriteBytesAt(dest, block[downTo-size:downTo])
			if err != nil {
				return err
			}
			break
		}

		size -= downTo
		err := r.WriteBytesAt(dest+size, block[0:downTo])
		if err != nil {
			return err
		}
		blockIndex--
		block, ok = r.blocks.Get(int(blockIndex))
		if !ok {
			return ErrItemNotFound
		}
		downTo = r.blockSize
	}
	return nil
}

// CopyBytesToArray Copies bytes from this store to a target byte array.
func (r *ByteStore) CopyBytesToArray(src int64, dest []byte) error {
	blockIndex := src >> r.blockBits
	upto := src & r.blockMask

	block, ok := r.blocks.Get(int(blockIndex))
	if !ok {
		return ErrItemNotFound
	}

	offset, size := int64(0), int64(len(dest))

	for size > 0 {
		chunk := r.blockSize - upto
		if size <= chunk {
			copy(dest[offset:offset+size], block[upto:])
			break
		}

		copy(dest[offset:offset+chunk], block[upto:])
		blockIndex++

		block, ok = r.blocks.Get(int(blockIndex))
		if !ok {
			return ErrItemNotFound
		}

		upto = 0
		size -= chunk
		offset += chunk
	}
	return nil
}

// WriteInt32 Writes an int at the absolute position without changing the current pointer.
func (r *ByteStore) WriteInt32(pos int64, value int32) error {
	blockIndex := int64(pos >> r.blockBits)
	upto := pos & r.blockMask
	block, ok := r.blocks.Get(int(blockIndex))
	if !ok {
		return ErrItemNotFound
	}
	shift := 24

	for i := 0; i < 4; i++ {
		block[upto] = (byte)(value >> shift)
		upto++
		shift -= 8
		if upto == r.blockSize {
			upto = 0
			blockIndex++
			block, ok = r.blocks.Get(int(blockIndex))
			if !ok {
				return ErrItemNotFound
			}
		}
	}
	return nil
}

// Reverse from srcPos, inclusive, to destPos, inclusive.
func (r *ByteStore) Reverse(srcPos, destPos int64) error {
	if srcPos > destPos {
		return errors.New("srcPos > destPos")
	}
	if destPos > r.GetPosition() {
		return errors.New("destPos bigger than position")
	}

	srcBlockIndex := srcPos >> r.blockBits
	src := srcPos & r.blockMask
	srcBlock, ok := r.blocks.Get(int(srcBlockIndex))
	if !ok {
		return ErrItemNotFound
	}

	destBlockIndex := destPos >> r.blockBits
	dest := destPos & r.blockMask
	destBlock, ok := r.blocks.Get(int(destBlockIndex))
	if !ok {
		return ErrItemNotFound
	}

	limit := (destPos - srcPos + 1) / 2

	for i := int64(0); i < limit; i++ {
		b := srcBlock[src]
		srcBlock[src] = destBlock[dest]
		destBlock[dest] = b
		src++

		if src == r.blockSize {
			srcBlockIndex++
			srcBlock, ok = r.blocks.Get(int(srcBlockIndex))
			if !ok {
				return ErrItemNotFound
			}
			src = 0
		}

		dest--

		if dest == -1 {
			destBlockIndex--
			destBlock, ok = r.blocks.Get(int(destBlockIndex))
			if !ok {
				return ErrItemNotFound
			}
			dest = r.blockSize - 1
		}
	}
	return nil
}

func (r *ByteStore) SkipBytes(size int64) error {
	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			r.nextWrite += size
			break
		}

		size -= chunk
		current := make([]byte, r.blockSize)
		r.blocks.Add(current)
		r.nextWrite = 0
	}
	return nil
}

func (r *ByteStore) GetPosition() int64 {
	return int64(r.blocks.Size()-1)*r.blockSize + r.nextWrite
}

// Truncate Pos must be less than the max position written so far! Ie, you cannot "grow" the file with this!
func (r *ByteStore) Truncate(newLen int64) error {
	if newLen > r.GetPosition() {
		return errors.New("newLen > r.GetPosition()")
	}

	if newLen < 0 {
		return errors.New("newLen < 0")
	}

	blockIndex := newLen >> r.blockBits
	nextWrite := newLen & r.blockMask
	if nextWrite == 0 {
		blockIndex--
		nextWrite = r.blockSize
	}

	r.blocks.RemoveRange(int(blockIndex+1), r.blocks.Size())

	if newLen == 0 {
		r.current = nil
	} else {
		var ok bool
		r.current, ok = r.blocks.Get(int(blockIndex))
		if !ok {
			return ErrItemNotFound
		}
	}

	if newLen != r.GetPosition() {
		return errors.New("newLen != r.GetPosition()")
	}

	return nil
}

func (r *ByteStore) Finish() error {
	if r.current != nil {
		// 减少内存消耗
		lastBuffer := make([]byte, r.nextWrite)
		copy(lastBuffer[:r.nextWrite], r.current[:r.nextWrite])
		r.blocks.Set(r.blocks.Size()-1, lastBuffer)
		r.current = nil
	}
	return nil
}

// WriteTo Writes all of our bytes to the target DataOutput.
func (r *ByteStore) WriteTo(out store.DataOutput) error {
	for _, block := range r.blocks.Values() {
		_, err := out.Write(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ByteStore) GetReverseReader() (BytesReader, error) {
	return r.getReverseReader(true)
}

func (r *ByteStore) getReverseReader(allowSingle bool) (BytesReader, error) {
	if allowSingle && r.blocks.Size() == 1 {
		item, ok := r.blocks.Get(0)
		if !ok {
			return nil, ErrItemNotFound
		}
		return NewReverseBytesReader(item), nil
	}

	return NewBuilderBytesReader(r)
}
