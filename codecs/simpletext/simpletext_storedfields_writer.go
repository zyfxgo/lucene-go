package simpletext

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"strconv"
)

var (
	FIELDS_EXTENSION         = "fld"
	STORED_FIELD_TYPE_STRING = []byte("string")
	STORED_FIELD_TYPE_BINARY = []byte("binary")
	STORED_FIELD_TYPE_INT    = []byte("int")
	STORED_FIELD_TYPE_LONG   = []byte("long")
	STORED_FIELD_TYPE_FLOAT  = []byte("float")
	STORED_FIELD_TYPE_DOUBLE = []byte("double")
	STORED_FIELD_END         = []byte("END")
	STORED_FIELD_DOC         = []byte("doc ")
	STORED_FIELD_FIELD       = []byte("  field ")
	STORED_FIELD_NAME        = []byte("    name ")
	STORED_FIELD_TYPE        = []byte("    type ")
	STORED_FIELD_VALUE       = []byte("    value ")
)

var _ index.StoredFieldsWriter = &SimpleTextStoredFieldsWriter{}

type SimpleTextStoredFieldsWriter struct {
	numDocsWritten int
	out            store.IndexOutput
	scratch        *bytes.Buffer
}

func newSimpleTextStoredFieldsWriter() *SimpleTextStoredFieldsWriter {
	return &SimpleTextStoredFieldsWriter{}
}

func NewSimpleTextStoredFieldsWriter(dir store.Directory,
	segment string, context *store.IOContext) (*SimpleTextStoredFieldsWriter, error) {
	writer := newSimpleTextStoredFieldsWriter()
	out, err := dir.CreateOutput(store.SegmentFileName(segment, "", FIELDS_EXTENSION), context)
	if err != nil {
		return nil, err
	}
	writer.out = out
	return writer, nil
}

func (s *SimpleTextStoredFieldsWriter) Close() error {
	return s.out.Close()
}

func (s *SimpleTextStoredFieldsWriter) StartDocument() error {
	if err := s.write(STORED_FIELD_DOC); err != nil {
		return err
	}
	if err := s.write(s.numDocsWritten); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	s.numDocsWritten++
	return nil
}

func (s *SimpleTextStoredFieldsWriter) FinishDocument() error {
	return nil
}

func (s *SimpleTextStoredFieldsWriter) WriteField(info *types.FieldInfo, field types.IndexableField) error {
	if err := s.write(STORED_FIELD_FIELD); err != nil {
		return err
	}
	if err := s.write(info.Number()); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}

	if err := s.write(STORED_FIELD_NAME); err != nil {
		return err
	}
	if err := s.write(info.Name()); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	if err := s.write(STORED_FIELD_TYPE); err != nil {
		return err
	}
	n := field.Value()

	if n != nil {
		switch n.(type) {
		case byte, int16, int32, int:
			return s.writeValue(STORED_FIELD_TYPE_INT, fmt.Sprintf("%v", n))
		case int64:
			return s.writeValue(STORED_FIELD_TYPE_LONG, fmt.Sprintf("%v", n))
		case float32:
			return s.writeValue(STORED_FIELD_TYPE_FLOAT, fmt.Sprintf("%v", n))
		case float64:
			return s.writeValue(STORED_FIELD_TYPE_DOUBLE, fmt.Sprintf("%v", n))
		case []byte:
			return s.writeValue(STORED_FIELD_TYPE_BINARY, string(n.([]byte)))
		case string:
			return s.writeValue(STORED_FIELD_TYPE_STRING, n.(string))
		default:
			return errors.New("cannot store numeric type")
		}
	}
	return errors.New("cannot store type")
}

func (s *SimpleTextStoredFieldsWriter) Finish(fis *index.FieldInfos, numDocs int) error {
	if s.numDocsWritten != numDocs {
		return errors.New("mergeFields produced an invalid result")
	}
	if err := s.write(STORED_FIELD_END); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	return utils.WriteChecksum(s.out)
}

func (s *SimpleTextStoredFieldsWriter) writeValue(valueType []byte, value string) error {
	if err := utils.WriteBytes(s.out, valueType); err != nil {
		return err
	}
	if err := utils.Newline(s.out); err != nil {
		return err
	}
	if err := utils.WriteBytes(s.out, STORED_FIELD_VALUE); err != nil {
		return err
	}

	if err := utils.WriteString(s.out, value); err != nil {
		return err
	}
	return utils.Newline(s.out)
}

func (s *SimpleTextStoredFieldsWriter) write(value any) error {
	switch value.(type) {
	case []byte:
		return utils.WriteBytes(s.out, value.([]byte))
	case string:
		return utils.WriteString(s.out, value.(string))
	case int:
		return utils.WriteString(s.out, strconv.Itoa(value.(int)))
	default:
		return nil
	}
}

func (s *SimpleTextStoredFieldsWriter) newLine() error {
	return utils.Newline(s.out)
}