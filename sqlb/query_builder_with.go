package sqlb

import "git.qjebbs.com/jebbs/go-sqls"

// With adds a segment as common table expression, the built query of s should be a subquery.
func (b *QueryBuilder) With(name, alias sqls.Table, s *sqls.Segment) *QueryBuilder {
	b.commonTableExprs = append(b.commonTableExprs, &namedSegment{
		Segment: s,
		name:    name,
		alias:   alias,
	})
	return b
}

type namedSegment struct {
	*sqls.Segment
	name, alias sqls.Table
}
