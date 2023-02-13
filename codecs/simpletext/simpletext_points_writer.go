package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/codecs/bkd"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.PointsWriter = &SimpleTextPointsWriter{}

var (
	NUM_DATA_DIMS   = []byte("num data dims ")
	NUM_INDEX_DIMS  = []byte("num index dims ")
	BYTES_PER_DIM   = []byte("bytes per dim ")
	MAX_LEAF_POINTS = []byte("max leaf points ")
	INDEX_COUNT     = []byte("index count ")
	BLOCK_COUNT     = []byte("block count ")
	BLOCK_DOC_ID    = []byte("  doc ")
	BLOCK_FP        = []byte("  block fp ")
	BLOCK_VALUE     = []byte("  block value ")
	SPLIT_COUNT     = []byte("split count ")
	SPLIT_DIM       = []byte("  split dim ")
	SPLIT_VALUE     = []byte("  split value ")
	FIELD_COUNT     = []byte("field count ")
	FIELD_FP_NAME   = []byte("  field fp name ")
	FIELD_FP        = []byte("  field fp ")
	MIN_VALUE       = []byte("min value ")
	MAX_VALUE       = []byte("max value ")
	POINT_COUNT     = []byte("point count ")
	DOC_COUNT       = []byte("doc count ")
	END             = []byte("END")
)

type SimpleTextPointsWriter struct {
	*index.PointsWriterDefault

	dataOut    store.IndexOutput
	scratch    *bytes.Buffer
	writeState *index.SegmentWriteState
	indexFPs   map[string]int64
}

func NewSimpleTextPointsWriter(writeState *index.SegmentWriteState) (*SimpleTextPointsWriter, error) {
	fileName := store.SegmentFileName(writeState.SegmentInfo.Name(), writeState.SegmentSuffix, POINT_EXTENSION)
	out, err := writeState.Directory.CreateOutput(fileName, writeState.Context)
	if err != nil {
		return nil, err
	}
	writer := &SimpleTextPointsWriter{
		PointsWriterDefault: nil,
		dataOut:             out,
		scratch:             new(bytes.Buffer),
		writeState:          writeState,
		indexFPs:            make(map[string]int64),
	}
	writer.PointsWriterDefault = &index.PointsWriterDefault{
		WriteField: writer.WriteField,
		Finish:     writer.Finish,
	}
	return writer, nil
}

func (s *SimpleTextPointsWriter) Close() error {
	if s.dataOut == nil {
		return nil
	}

	if err := s.dataOut.Close(); err != nil {
		return err
	}
	s.dataOut = nil

	fileName := store.SegmentFileName(s.writeState.SegmentInfo.Name(),
		s.writeState.SegmentSuffix, POINT_INDEX_EXTENSION)

	indexOut, err := s.writeState.Directory.CreateOutput(fileName, s.writeState.Context)
	if err != nil {
		return err
	}
	count := len(s.indexFPs)

	w := utils.NewTextWriter(indexOut)
	w.WriteBytes(FIELD_COUNT)
	w.WriteInt(count)
	w.NewLine()

	for k, v := range s.indexFPs {
		w.WriteBytes(FIELD_FP_NAME)
		w.WriteString(k)
		w.NewLine()

		w.WriteBytes(FIELD_FP)
		w.WriteInt(int(v))
		w.NewLine()
	}
	return w.Checksum()
}

func (s *SimpleTextPointsWriter) WriteField(fieldInfo *types.FieldInfo, reader index.PointsReader) error {
	values, err := reader.GetValues(fieldInfo.Name())
	if err != nil {
		return err
	}

	config, err := bkd.NewBKDConfig(
		fieldInfo.GetPointDimensionCount(),
		fieldInfo.GetPointIndexDimensionCount(),
		fieldInfo.GetPointNumBytes(),
		bkd.DEFAULT_MAX_POINTS_IN_LEAF_NODE,
	)
	if err != nil {
		return err
	}

	maxDoc, err := s.writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return err
	}
	writer := NewSimpleTextBKDWriter(maxDoc,
		s.writeState.Directory,
		s.writeState.SegmentInfo.Name(),
		config,
		DEFAULT_MAX_MB_SORT_IN_HEAP,
		values.Size())

	err = values.Intersect(&index.IntersectVisitor{
		VisitByDocID: func(docID int) error {
			return errors.New("illegal State")
		},
		VisitLeaf: func(docID int, packedValue []byte) error {
			return writer.Add(packedValue, docID)
		},
		Compare: func(minPackedValue, maxPackedValue []byte) index.Relation {
			return index.CELL_CROSSES_QUERY
		},
		Grow: func(count int) {
		},
	})
	if err != nil {
		return err
	}

	// We could have 0 points on merge since all docs with points may be deleted:
	if writer.GetPointCount() > 0 {
		fp, err := writer.Finish(s.dataOut)
		if err != nil {
			return err
		}
		s.indexFPs[fieldInfo.Name()] = fp
	}

	return s.dataOut.Close()
}

func (s *SimpleTextPointsWriter) Finish() error {
	if err := utils.WriteBytes(s.dataOut, END); err != nil {
		return err
	}
	if err := utils.WriteNewline(s.dataOut); err != nil {
		return err
	}
	return utils.WriteChecksum(s.dataOut)
}
