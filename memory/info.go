package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

type Info struct {
	index *MemoryIndex

	fieldInfo *types.FieldInfo
	norm      *int64

	// TODO
	// Term strings and their positions for this field: Map <String termText, ArrayIntList positions>
	// private BytesHash terms;
	terms *util.BytesHash
	// private SliceByteStartArray sliceArray;
	sliceArray *SliceByteStartArray

	// Terms sorted ascending by term text; computed on demand
	sortedTerms []int

	// Number of added tokens for this field
	numTokens int

	// Number of overlapping tokens for this field
	numOverlapTokens int

	sumTotalTermFreq int64

	maxTermFrequency int

	// the last position encountered in this field for multi field support
	lastPosition int

	// the last offset encountered in this field for multi field support
	lastOffset int

	binaryProducer  *BinaryDocValuesProducer
	numericProducer *NumericDocValuesProducer

	preparedDocValuesAndPointValues bool

	pointValues [][]byte

	minPackedValue   []byte
	maxPackedValue   []byte
	pointValuesCount int
}

func (m *MemoryIndex) NewInfo(fieldInfo *types.FieldInfo, byteBlockPool *util.ByteBlockPool) *Info {
	sliceArray := NewSliceByteStartArray(util.DEFAULT_CAPACITY)

	info := Info{
		index:           m,
		fieldInfo:       fieldInfo,
		terms:           util.NewBytesRefHashV1(byteBlockPool, util.DEFAULT_CAPACITY, sliceArray),
		sliceArray:      sliceArray,
		sortedTerms:     make([]int, 0),
		binaryProducer:  NewBinaryDocValuesProducer(),
		numericProducer: NewNumericDocValuesProducer(),
		pointValues:     make([][]byte, 0),
		minPackedValue:  make([]byte, 0),
		maxPackedValue:  make([]byte, 0),
	}

	return &info
}

func (r *Info) freeze() {

}

// Sorts hashed Terms into ascending order, reusing memory along the way. Note that sorting is lazily
// delayed until required (often it's not required at all). If a sorted view is required then
// hashing + sort + binary search is still faster and smaller than TreeMap usage (which would be an
// alternative and somewhat more elegant approach, apart from more sophisticated Tries / prefix trees).
func (r *Info) sortTerms() {
	if len(r.sortedTerms) == 0 {
		r.sortedTerms = r.terms.Sort()
	}
}

func (r *Info) prepareDocValuesAndPointValues() {

}

func (r *Info) getNormDocValues() index.NumericDocValues {
	if r.norm == nil {
		invertState := index.NewFieldInvertState(
			util.VersionLast.Major,
			r.fieldInfo.Name(),
			r.fieldInfo.GetIndexOptions(),
			r.lastPosition,
			r.numTokens,
			r.numOverlapTokens,
			0,
			r.maxTermFrequency,
			r.terms.Size())

		value := r.index.normSimilarity.ComputeNorm(invertState)
		r.norm = &value

		return newInnerNumericDocValues(*r.norm)
	}

	//if (norm == null) {
	//	FieldInvertState invertState =
	//		new FieldInvertState(
	//		Version.LATEST.major,
	//		fieldInfo.name,
	//		fieldInfo.getIndexOptions(),
	//		lastPosition,
	//		numTokens,
	//		numOverlapTokens,
	//		0,
	//		maxTermFrequency,
	//		terms.size());
	//	final long value = normSimilarity.computeNorm(invertState);
	//	if (DEBUG)
	//		System.err.println(
	//			"MemoryIndexReader.norms: " + fieldInfo.name + ":" + value + ":" + numTokens);
	//
	//	norm = value;
	// TODO
	panic("TODO")
}
