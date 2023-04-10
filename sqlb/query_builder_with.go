package sqlb

import "github.com/qjebbs/go-sqls"

// With adds a segment as common table expression, the built query of s should be a subquery.
func (b *QueryBuilder) With(name sqls.Table, s *sqls.Segment) *QueryBuilder {
	b.ctes = append(b.ctes, &cte{
		table:   NewTable(name, ""),
		Segment: s,
	})
	return b
}

type cte struct {
	table Table
	*sqls.Segment
}
