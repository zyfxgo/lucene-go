package index

import (
	"github.com/geange/lucene-go/core/types"
	"sort"
	"strings"
)

type FieldData struct {
	fieldInfo     *types.FieldInfo
	terms         []*TermData
	omitTF        bool
	storePayloads bool
}

func NewFieldData(name string, fieldInfos *FieldInfosBuilder, terms []*TermData, omitTF, storePayloads bool) (*FieldData, error) {
	data := &FieldData{
		omitTF:        omitTF,
		storePayloads: storePayloads,
	}
	// TODO: change this test to use all three
	fieldInfo, err := fieldInfos.GetOrAdd(name)
	if err != nil {
		return nil, err
	}

	if omitTF {
		err := fieldInfo.SetIndexOptions(types.INDEX_OPTIONS_DOCS)
		if err != nil {
			return nil, err
		}
	} else {
		err := fieldInfo.SetIndexOptions(types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
		if err != nil {
			return nil, err
		}
	}

	if storePayloads {
		err := fieldInfo.SetStorePayloads()
		if err != nil {
			return nil, err
		}
	}
	data.fieldInfo = fieldInfo
	data.terms = terms
	for i := range terms {
		terms[i].field = data
	}
	sort.Sort(TermDataList(data.terms))
	return data, nil
}

type FieldDataList []FieldData

func (f FieldDataList) Len() int {
	return len(f)
}

func (f FieldDataList) Less(i, j int) bool {
	return strings.Compare(f[i].fieldInfo.Name(), f[j].fieldInfo.Name()) < 0
}

func (f FieldDataList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type TermData struct {
	text2     string
	text      []byte
	docs      []int
	positions [][]PositionData
	field     *FieldData
}

type TermDataList []*TermData

func (t TermDataList) Len() int {
	return len(t)
}

func (t TermDataList) Less(i, j int) bool {
	return strings.Compare(t[i].text2, t[j].text2) < 0
}

func (t TermDataList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func NewTermData(text string, docs []int, positions [][]PositionData) *TermData {
	return &TermData{
		text2:     text,
		text:      []byte(text),
		docs:      docs,
		positions: positions,
	}
}

type PositionData struct {
	Pos     int
	Payload []byte
}

func NewPositionData(pos int, payload []byte) *PositionData {
	return &PositionData{
		Pos:     pos,
		Payload: payload,
	}
}