package search

var _ Scorable = &ScoreAndDoc{}

type ScoreAndDoc struct {
	*ScorableBase

	score float64
	doc   int
}

func NewScoreAndDoc() *ScoreAndDoc {
	return &ScoreAndDoc{ScorableBase: &ScorableBase{}}
}

func (s *ScoreAndDoc) Score() (float64, error) {
	return s.score, nil
}

func (s *ScoreAndDoc) DocID() int {
	return s.doc
}
