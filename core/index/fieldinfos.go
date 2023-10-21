package index

import (
	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/document"
)

// FieldInfos Collection of FieldInfos (accessible by number or by name).
type FieldInfos struct {
	hasFreq          bool
	hasProx          bool
	hasPayloads      bool
	hasOffsets       bool
	hasVectors       bool
	hasNorms         bool
	hasDocValues     bool
	hasPointValues   bool
	softDeletesField string

	// used only by fieldInfo(int)
	byNumber []*document.FieldInfo

	byName map[string]*document.FieldInfo
	values []*document.FieldInfo // for an unmodifiable iterator
}

func NewFieldInfos(infos []*document.FieldInfo) *FieldInfos {
	hasVectors := false
	hasProx := false
	hasPayloads := false
	hasOffsets := false
	hasFreq := false
	hasNorms := false
	hasDocValues := false
	hasPointValues := false
	softDeletesField := ""

	tmap := treeset.NewWith[*document.FieldInfo](func(info1, info2 *document.FieldInfo) int {
		if info1.Number() == info2.Number() {
			return 0
		} else if info1.Number() > info2.Number() {
			return 1
		} else {
			return -1
		}
	})

	maxNum := 0
	for _, info := range infos {
		if info.Number() > maxNum {
			maxNum = info.Number()
		}
	}

	this := &FieldInfos{
		byName:   map[string]*document.FieldInfo{},
		byNumber: []*document.FieldInfo{},
	}

	for _, info := range infos {
		if info.Number() < 0 {
			panic("")
		}

		if tmap.Contains(info) {
			panic("")
		}

		tmap.Add(info)

		if _, ok := this.byName[info.Name()]; ok {
			panic("")
		} else {
			this.byName[info.Name()] = info
		}

		hasVectors = hasVectors || info.HasVectors()
		hasProx = hasProx || info.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
		hasFreq = hasFreq || info.GetIndexOptions() != document.INDEX_OPTIONS_DOCS
		hasOffsets = hasOffsets || info.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		hasNorms = hasNorms || info.HasNorms()
		hasDocValues = hasDocValues || info.GetDocValuesType() != document.DOC_VALUES_TYPE_NONE
		hasPayloads = hasPayloads || info.HasPayloads()
		hasPointValues = hasPointValues || info.GetPointDimensionCount() != 0

		if info.IsSoftDeletesField() {
			if softDeletesField == info.Name() {
				panic("")
			}
			softDeletesField = info.Name()
		}
	}

	this.hasVectors = hasVectors
	this.hasProx = hasProx
	this.hasPayloads = hasPayloads
	this.hasOffsets = hasOffsets
	this.hasFreq = hasFreq
	this.hasNorms = hasNorms
	this.hasDocValues = hasDocValues
	this.hasPointValues = hasPointValues
	this.softDeletesField = softDeletesField

	values := tmap.Values()
	items := make([]*document.FieldInfo, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	this.byNumber = items
	this.values = items

	return this
}

func (f *FieldInfos) FieldInfo(fieldName string) *document.FieldInfo {
	return f.byName[fieldName]
}

func (f *FieldInfos) FieldInfoByNumber(fieldNumber int) *document.FieldInfo {
	return f.byNumber[fieldNumber]
}

func (f *FieldInfos) Size() int {
	return len(f.byName)
}

func (f *FieldInfos) List() []*document.FieldInfo {
	return f.values
}

func (f *FieldInfos) HasNorms() bool {
	return f.hasNorms
}

func (f *FieldInfos) HasDocValues() bool {
	return f.hasDocValues
}

func (f *FieldInfos) HasVectors() bool {
	return f.hasVectors
}

func (f *FieldInfos) HasPointValues() bool {
	return f.hasPointValues
}