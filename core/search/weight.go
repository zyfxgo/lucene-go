package search

import "github.com/geange/lucene-go/core/index"

// Weight Expert: Calculate query weights and build query scorers.
// 计算查询权重并构建查询记分器。
// The purpose of Weight is to ensure searching does not modify a Query, so that a Query instance can be reused.
// IndexSearcher dependent state of the query should reside in the Weight. LeafReader dependent state should
// reside in the Scorer.
// 权重的目的是确保搜索不会修改查询，以便可以重用查询实例。查询的IndexSearcher依赖状态应位于权重中。LeafReader 相关状态应位于记分器中。
//
// Since Weight creates Scorer instances for a given LeafReaderContext (scorer(LeafReaderContext)) callers must
// maintain the relationship between the searcher's top-level IndexReaderContext and the context used to create
//
// 由于权重为给定的LeafReaderContext（Scorer（LeafReaderContext））创建记分器实例，因此调用程序必须保持搜索器的顶级索引
// ReaderContext和用于创建记分器的上下文之间的关系。
// a Scorer.
// A Weight is used in the following way:
// A Weight is constructed by a top-level query, given a IndexSearcher (Query.createWeight(IndexSearcher, ScoreMode, float)).
// A Scorer is constructed by scorer(LeafReaderContext).
// Since: 2.9
type Weight interface {
	// Matches Returns Matches for a specific document, or null if the document does not match the parent query A query match that contains no position information (for example, a Point or DocValues query) will return MatchesUtils.MATCH_WITH_NO_TERMS
	// Params: 	context – the reader's context to create the Matches for
	//			doc – the document's id relative to the given context's reader
	Matches(context *index.LeafReaderContext, doc int) (Matches, error)

	// Explain An explanation of the score computation for the named document.
	// Params: 	context – the readers context to create the Explanation for.
	//			doc – the document's id relative to the given context's reader
	// Returns: an Explanation for the score
	// Throws: 	IOException – if an IOException occurs
	Explain(ctx *index.LeafReaderContext, doc int) (*Explanation, error)

	// GetQuery The query that this concerns.
	GetQuery() Query

	// Scorer Returns a Scorer which can iterate in order over all matching documents and assign them a score.
	//NOTE: null can be returned if no documents will be scored by this query.
	//NOTE: The returned Scorer does not have LeafReader.getLiveDocs() applied, they need to be checked on top.
	//Params:
	//context – the LeafReaderContext for which to return the Scorer.
	//Returns:
	//a Scorer which scores documents in/out-of order.
	//Throws:
	//IOException – if there is a low-level I/O error
	Scorer(ctx *index.LeafReaderContext) (Scorer, error)

	// ScorerSupplier Optional method. Get a ScorerSupplier, which allows to know the cost of the Scorer before building it. The default implementation calls scorer and builds a ScorerSupplier wrapper around it.
	//See Also:
	//scorer
	ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error)

	// BulkScorer Optional method, to return a BulkScorer to score the query and send hits to a Collector. Only queries that have a different top-level approach need to override this; the default implementation pulls a normal Scorer and iterates and collects the resulting hits which are not marked as deleted.
	// Params: 	context – the LeafReaderContext for which to return the Scorer.
	// Returns: a BulkScorer which scores documents and passes them to a collector.
	// Throws: 	IOException – if there is a low-level I/O error
	BulkScorer(ctx *index.LeafReaderContext) (BulkScorer, error)
}

type WeightExtra interface {
	Scorer(ctx *index.LeafReaderContext) (Scorer, error)
}

type WeightImp struct {
	WeightExtra

	parentQuery Query
}

func newWeightImp(parentQuery Query) *WeightImp {
	return &WeightImp{parentQuery: parentQuery}
}

func (r *WeightImp) Matches(ctx *index.LeafReaderContext, doc int) (Matches, error) {
	scorerSupplier, err := r.ScorerSupplier(ctx)
	if err != nil {
		return nil, err
	}
	if scorerSupplier == nil {
		return nil, nil
	}

	scorer, err := scorerSupplier.Get(1)
	if err != nil {
		return nil, err
	}
	twoPhase := scorer.TwoPhaseIterator()
	if twoPhase == nil {
		advance, err := scorer.Iterator().Advance(doc)
		if err != nil {
			return nil, err
		}
		if advance != doc {
			return nil, nil
		}
	} else {
		advance, err := twoPhase.Approximation().Advance(doc)
		if err != nil {
			return nil, err
		}

		if ok, _ := twoPhase.Matches(); advance != doc || !ok {
			return nil, nil
		}
	}
	panic("")
}

func (r *WeightImp) ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error) {
	scorer, err := r.Scorer(ctx)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, nil
	}

	return &scorerSupplier{scorer: scorer}, nil
}

var _ ScorerSupplier = &scorerSupplier{}

type scorerSupplier struct {
	scorer Scorer
}

func (s *scorerSupplier) Get(leadCost int64) (Scorer, error) {
	return s.scorer, nil
}

func (s *scorerSupplier) Cost() int64 {
	return s.scorer.Iterator().Cost()
}

func (r *WeightImp) BulkScorer(ctx *index.LeafReaderContext) (BulkScorer, error) {
	panic("")
}
