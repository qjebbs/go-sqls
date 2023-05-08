package sqlb

import "github.com/qjebbs/go-sqls"

// With adds a segment as common table expression, the built query of s should be a subquery.
func (b *QueryBuilder) With(name sqls.Table, builder sqls.Builder) *QueryBuilder {
	b.ctes = append(b.ctes, &cte{
		table:   NewTable(name, ""),
		Builder: builder,
	})
	return b
}

type cte struct {
	table Table
	sqls.Builder
}
