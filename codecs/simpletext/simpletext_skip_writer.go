package simpletext

import (
	"fmt"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	skipMultiplier = 3
	maxSkipLevels  = 4

	BLOCK_SIZE = 8
)

var (
	SKIP_LIST     = []byte("    skipList ")
	LEVEL_LENGTH  = []byte("      levelLength ")
	LEVEL         = []byte("      level ")
	SKIP_DOC      = []byte("        skipDoc ")
	SKIP_DOC_FP   = []byte("        skipDocFP ")
	IMPACTS       = []byte("        impacts ")
	IMPACT        = []byte("          impact ")
	FREQ          = []byte("            freq ")
	NORM          = []byte("            norm ")
	IMPACTS_END   = []byte("        impactsEnd ")
	CHILD_POINTER = []byte("        childPointer ")
)

var _ index.MultiLevelSkipListWriter = &SimpleTextSkipWriter{}

type SimpleTextSkipWriter struct {
	*index.MultiLevelSkipListWriterDefault

	wroteHeaderPerLevelMap  map[int]bool
	curDoc                  int
	curDocFilePointer       int64
	curCompetitiveFreqNorms []*index.CompetitiveImpactAccumulator
}

func NewSimpleTextSkipWriter(writeState *index.SegmentWriteState) (*SimpleTextSkipWriter, error) {
	maxDoc, err := writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	writer := &SimpleTextSkipWriter{
		wroteHeaderPerLevelMap:  make(map[int]bool),
		curDoc:                  0,
		curDocFilePointer:       0,
		curCompetitiveFreqNorms: make([]*index.CompetitiveImpactAccumulator, maxSkipLevels),
	}

	for i := range writer.curCompetitiveFreqNorms {
		writer.curCompetitiveFreqNorms[i] = index.NewCompetitiveImpactAccumulator()
	}

	writer.MultiLevelSkipListWriterDefault = index.NewMultiLevelSkipListWriterDefault(&index.MultiLevelSkipListWriterDefaultConfig{
		SkipInterval:      BLOCK_SIZE,
		SkipMultiplier:    skipMultiplier,
		MaxSkipLevels:     maxSkipLevels,
		DF:                maxDoc,
		WriteSkipData:     writer.WriteSkipData,
		WriteLevelLength:  writer.WriteLevelLength,
		WriteChildPointer: writer.WriteChildPointer,
	})
	writer.ResetSkip()
	return writer, nil
}

func (s *SimpleTextSkipWriter) WriteSkipData(level int, skipBuffer store.IndexOutput) error {
	wroteHeader := s.wroteHeaderPerLevelMap[level]

	w := utils.NewTextWriter(skipBuffer)

	if !wroteHeader {
		w.WriteBytes(LEVEL)
		w.WriteString(fmt.Sprintf("%d", level))
		w.NewLine()

		s.wroteHeaderPerLevelMap[level] = true
	}
	w.WriteBytes(SKIP_DOC)
	w.WriteString(fmt.Sprintf("%d", s.curDoc))
	w.NewLine()

	w.WriteBytes(SKIP_DOC_FP)
	w.WriteString(fmt.Sprintf("%d", s.curDocFilePointer))
	w.NewLine()

	competitiveFreqNorms := s.curCompetitiveFreqNorms[level]
	impacts := competitiveFreqNorms.GetCompetitiveFreqNormPairs()
	//assert impacts.size() > 0;
	if level+1 < s.NumberOfSkipLevels {
		s.curCompetitiveFreqNorms[level+1].AddAll(competitiveFreqNorms)
	}
	w.WriteBytes(IMPACTS)
	w.NewLine()
	for _, impact := range impacts {
		w.WriteBytes(IMPACT)
		w.NewLine()

		w.WriteBytes(FREQ)
		w.WriteString(fmt.Sprintf("%d", impact.Freq))
		w.NewLine()

		w.WriteBytes(NORM)
		w.WriteString(fmt.Sprintf("%d", impact.Norm))
		w.NewLine()
	}
	w.WriteBytes(IMPACTS_END)
	w.NewLine()
	competitiveFreqNorms.Clear()

	return nil
}

func (s *SimpleTextSkipWriter) ResetSkip() {
	s.MultiLevelSkipListWriterDefault.ResetSkip()
	s.wroteHeaderPerLevelMap = map[int]bool{}
	s.curDoc = -1
	s.curDocFilePointer = -1
	for _, norm := range s.curCompetitiveFreqNorms {
		norm.Clear()
	}
}

func (s *SimpleTextSkipWriter) WriteSkip(output store.IndexOutput) (int64, error) {
	skipOffset := output.GetFilePointer()

	w := utils.NewTextWriter(output)

	if err := w.WriteBytes(SKIP_LIST); err != nil {
		return 0, err
	}
	if err := w.NewLine(); err != nil {
		return 0, err
	}
	if _, err := s.MultiLevelSkipListWriterDefault.WriteSkip(output); err != nil {
		return 0, err
	}
	return skipOffset, nil
}

//func (s *SimpleTextSkipWriter) BufferSkipV1(doc int, docFilePointer int64, numDocs int, competitiveImpactAccumulator *index.CompetitiveImpactAccumulator) error {
//	//assert doc > curDoc;
//	s.curDoc = doc
//	s.curDocFilePointer = docFilePointer
//	s.curCompetitiveFreqNorms[0].AddAll(competitiveImpactAccumulator)
//	return s.BufferSkip(numDocs)
//}

func (s *SimpleTextSkipWriter) WriteLevelLength(levelLength int64, output store.IndexOutput) error {
	utils.WriteBytes(output, LEVEL_LENGTH)
	utils.WriteString(output, fmt.Sprintf("%d", levelLength))
	utils.Newline(output)
	return nil
}

func (s *SimpleTextSkipWriter) WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error {
	utils.WriteBytes(skipBuffer, CHILD_POINTER)
	utils.WriteString(skipBuffer, fmt.Sprintf("%d", childPointer))
	utils.Newline(skipBuffer)
	return nil
}

func (s *SimpleTextSkipWriter) bufferSkip(doc int, docFilePointer int64, numDocs int, accumulator *index.CompetitiveImpactAccumulator) error {
	s.curDoc = doc
	s.curDocFilePointer = docFilePointer
	s.curCompetitiveFreqNorms[0].AddAll(accumulator)
	return s.BufferSkip(numDocs)
}