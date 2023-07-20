package search

import (
	"github.com/geange/lucene-go/core/index"
)

var _ Query = &MatchNoDocsQuery{}

// MatchNoDocsQuery
// A query that matches no documents.
type MatchNoDocsQuery struct {
	reason string
}

func NewMatchNoDocsQuery() *MatchNoDocsQuery {
	return &MatchNoDocsQuery{}
}

func NewMatchNoDocsQueryV1(reason string) *MatchNoDocsQuery {
	return &MatchNoDocsQuery{reason: reason}
}

func (m *MatchNoDocsQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Rewrite(reader index.Reader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Visit(visitor QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}
