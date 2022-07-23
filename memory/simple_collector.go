package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
)

var _ search.SimpleCollector = &simpleCollector{}

type simpleCollector struct {
	*search.SimpleCollectorImp
	scorer search.Scorable
	scores []float64
}

func newSimpleCollector(scores []float64) *simpleCollector {
	collector := &simpleCollector{
		SimpleCollectorImp: nil,
		scorer:             nil,
		scores:             scores,
	}
	collector.SimpleCollectorExtra = collector
	return collector
}

func (s *simpleCollector) ScoreMode() *search.ScoreMode {
	return search.COMPLETE
}

func (s *simpleCollector) Collect(doc int) error {
	var err error
	s.scores[0], err = s.scorer.Score()
	return err
}

func (s *simpleCollector) DoSetNextReader(context *index.LeafReaderContext) error {
	return nil
}

func (s *simpleCollector) SetScorer(scorer search.Scorable) error {
	s.scorer = scorer
	return nil
}