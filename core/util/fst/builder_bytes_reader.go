package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &BuilderBytesReader{}

type BuilderBytesReader struct {
	*store.DataInputImp
	bs         *ByteStore
	current    []byte
	nextBuffer int
	nextRead   int
}

func NewBuilderBytesReader(bs *ByteStore) (*BuilderBytesReader, error) {
	var current []byte
	if bs.blocks.Size() != 0 {
		v, err := bs.blocks.Get(0)
		if err != nil {
			return nil, err
		}
		current = v
	}

	return &BuilderBytesReader{
		current:    current,
		bs:         bs,
		nextBuffer: -1,
		nextRead:   0,
	}, nil
}

func (b *BuilderBytesReader) ReadByte() (byte, error) {
	if b.nextRead == -1 {
		var err error
		b.current, err = b.bs.blocks.Get(b.nextBuffer)
		if err != nil {
			return 0, err
		}
		b.nextBuffer++
		b.nextRead = int(b.bs.blockSize - 1)
	}
	v := b.current[b.nextRead]
	b.nextRead--
	return v, nil
}

func (b *BuilderBytesReader) ReadBytes(bs []byte) error {
	for i := range bs {
		v, err := b.ReadByte()
		if err != nil {
			return err
		}
		bs[i] = v
	}
	return nil
}

func (b *BuilderBytesReader) GetPosition() int64 {
	return int64(b.nextBuffer+1)*b.bs.blockSize + int64(b.nextRead)
}

func (b *BuilderBytesReader) SetPosition(pos int64) error {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// EOF)...?
	bufferIndex := (int)(pos >> b.bs.blockBits)
	if b.nextBuffer != bufferIndex-1 {
		b.nextBuffer = bufferIndex - 1
		v, err := b.bs.blocks.Get(bufferIndex)
		if err != nil {
			return err
		}
		b.current = v
	}
	b.nextRead = int(pos & b.bs.blockMask)
	// TODO: assert getPosition() == pos : "pos=" + pos + " getPos()=" + getPosition();
	return nil
}

func (b *BuilderBytesReader) Reversed() bool {
	return true
}
