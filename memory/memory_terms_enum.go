package memory

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattr"
	"github.com/geange/lucene-go/core/util"
)

var _ index.TermsEnum = &MemoryTermsEnum{}

type MemoryTermsEnum struct {
	info     *Info
	termUpto int
	br       []byte
	atts     *tokenattr.AttributeSource

	*MemoryIndex
}

func (m *MemoryIndex) NewMemoryTermsEnum(info *Info) *MemoryTermsEnum {
	info.sortTerms()
	return &MemoryTermsEnum{
		info:        info,
		termUpto:    -1,
		br:          nil,
		atts:        tokenattr.NewAttributeSource(),
		MemoryIndex: m,
	}
}

func (m *MemoryTermsEnum) binarySearch(b []byte, low, high int, hash *util.BytesHash, ords []int) int {
	mid := 0
	for low <= high {
		mid = (low + high) >> 1
		bytesRef := hash.Get(ords[mid])

		cmp := bytes.Compare(bytesRef, b)
		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid
		}
	}
	return -(low + 1)
}

func (m *MemoryTermsEnum) Next() ([]byte, error) {
	m.termUpto++
	if m.termUpto >= m.info.terms.Size() {
		return nil, nil
	}
	m.br = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
	return m.br, nil
}

func (m *MemoryTermsEnum) Attributes() *tokenattr.AttributeSource {
	return m.atts
}

func (m *MemoryTermsEnum) SeekExact(text []byte) (bool, error) {
	m.termUpto = m.binarySearch(text, 0, m.info.terms.Size(), m.info.terms, m.info.sortedTerms)
	return m.termUpto >= 0, nil
}

func (m *MemoryTermsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	m.termUpto = m.binarySearch(text, 0, m.info.terms.Size()-1, m.info.terms, m.info.sortedTerms)
	if m.termUpto < 0 { // not found; choose successor
		m.termUpto = -m.termUpto - 1
		if m.termUpto >= m.info.terms.Size() {
			return index.SEEK_STATUS_END, nil
		} else {
			m.br = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
			return index.SEEK_STATUS_NOT_FOUND, nil
		}
	} else {
		return index.SEEK_STATUS_FOUND, nil
	}
}

func (m *MemoryTermsEnum) SeekExactByOrd(ord int64) error {
	m.termUpto = int(ord)
	m.br = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
	return nil
}

func (m *MemoryTermsEnum) SeekExactExpert(term []byte, state index.TermState) error {
	return m.SeekExactByOrd(state.(*index.OrdTermState).Ord)
}

func (m *MemoryTermsEnum) Term() ([]byte, error) {
	return m.br, nil
}

func (m *MemoryTermsEnum) Ord() (int64, error) {
	return int64(m.termUpto), nil
}

func (m *MemoryTermsEnum) DocFreq() (int, error) {
	return 1, nil
}

func (m *MemoryTermsEnum) TotalTermFreq() (int64, error) {
	return int64(m.info.sliceArray.freq[m.info.sortedTerms[m.termUpto]]), nil
}

func (m *MemoryTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	if reuse == nil {
		reuse = m.NewMemoryPostingsEnum()
	}

	if _, ok := reuse.(*MemoryPostingsEnum); !ok {
		reuse = m.NewMemoryPostingsEnum()
	}

	ord := m.info.sortedTerms[m.termUpto]

	return reuse.(*MemoryPostingsEnum).reset(m.info.sliceArray.start[ord],
		m.info.sliceArray.end[ord], m.info.sliceArray.freq[ord]), nil
}

func (m *MemoryTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	postings, err := m.Postings(nil, flags)
	if err != nil {
		return nil, err
	}
	return index.NewSlowImpactsEnum(postings), nil
}

func (m *MemoryTermsEnum) TermState() (index.TermState, error) {
	ts := index.NewOrdTermState()
	ts.Ord = int64(m.termUpto)
	return ts, nil
}
