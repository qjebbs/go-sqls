package sqlb

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

// InnerJoin append a inner join table.
func (b *QueryBuilder) InnerJoin(name, alias sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("INNER JOIN", name, alias, on, false)
}

// LeftJoin append / replace a left join table.
func (b *QueryBuilder) LeftJoin(name, alias sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("LEFT JOIN", name, alias, on, false)
}

// LeftJoinOptional append / replace a left join table, and mark it as optional.
//
// CAUSION:
//
//   - Make sure all columns referenced in the query are reflected in
//     *sqls.Segment.Columns, so that the *QueryBuilder can calculate the dependency
//     correctly.
//   - Make sure it's used with the SELECT DISTINCT statement, otherwise it works
//     exactly the same as LeftJoin().
//
// Consider the following two queries:
//
//	SELECT DISTINCT foo.* FROM foo LEFT JOIN bar ON foo.id = bar.foo_id
//	SELECT DISTINCT foo.* FROM foo
//
// They return the same result, but the second query more efficient.
// If the join to "bar" is declared with LeftJoinOptional(), *QueryBuilder
// will trim it if no relative columns referenced in the query, aka Join Elimination.
func (b *QueryBuilder) LeftJoinOptional(name, alias sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("LEFT JOIN", name, alias, on, true)
}

// RightJoin append / replace a right join table.
func (b *QueryBuilder) RightJoin(name, alias sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("RIGHT JOIN", name, alias, on, false)
}

// FullJoin append / replace a full join table.
func (b *QueryBuilder) FullJoin(name, alias sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("FULL JOIN", name, alias, on, false)
}

// CrossJoin append / replace a cross join table.
func (b *QueryBuilder) CrossJoin(name, alias sqls.Table) *QueryBuilder {
	return b.join("CROSS JOIN", name, alias, nil, false)
}

// From append or replace a from table.
func (b *QueryBuilder) join(joinStr string, name, alias sqls.Table, on *sqls.Segment, optional bool) *QueryBuilder {
	if name == "" || alias == "" {
		b.pushError(fmt.Errorf("join table name or alias is empty"))
		return b
	}
	if _, ok := b.tablesByName[alias]; !ok {
		if len(b.tableNames) == 0 {
			// reserve the first alias for the main table
			b.tableNames = append(b.tableNames, "")
		}
		b.tableNames = append(b.tableNames, alias)
	}
	tableAndAlias := name
	if alias != "" {
		tableAndAlias = tableAndAlias + " AS " + alias
	}
	if on == nil || on.Raw == "" {
		b.tablesByName[alias] = &fromTable{
			Table: alias,
			Segment: &sqls.Segment{
				Raw: fmt.Sprintf("%s %s", joinStr, tableAndAlias),
			},
			Optional: optional,
		}
		return b
	}
	b.tablesByName[alias] = &fromTable{
		Table: alias,
		Segment: &sqls.Segment{
			Raw:      fmt.Sprintf("%s %s ON %s", joinStr, tableAndAlias, on.Raw),
			Segments: on.Segments,
			Columns:  on.Columns,
			Args:     on.Args,
		},
		Optional: optional,
	}
	return b
}
