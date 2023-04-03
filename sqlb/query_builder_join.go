package sqlb

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

// InnerJoin append a inner join table.
func (b *QueryBuilder) InnerJoin(t sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("INNER JOIN", t, on, false)
}

// LeftJoin append / replace a left join table.
func (b *QueryBuilder) LeftJoin(t sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("LEFT JOIN", t, on, false)
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
func (b *QueryBuilder) LeftJoinOptional(t sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("LEFT JOIN", t, on, true)
}

// RightJoin append / replace a right join table.
func (b *QueryBuilder) RightJoin(t sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("RIGHT JOIN", t, on, false)
}

// FullJoin append / replace a full join table.
func (b *QueryBuilder) FullJoin(t sqls.Table, on *sqls.Segment) *QueryBuilder {
	return b.join("FULL JOIN", t, on, false)
}

// CrossJoin append / replace a cross join table.
func (b *QueryBuilder) CrossJoin(t sqls.Table) *QueryBuilder {
	return b.join("CROSS JOIN", t, nil, false)
}

// From append or replace a from table.
func (b *QueryBuilder) join(joinStr string, table sqls.Table, on *sqls.Segment, optional bool) *QueryBuilder {
	if table[0] == "" {
		b.pushError(fmt.Errorf("join table name is empty"))
		return b
	}
	if _, ok := b.tablesByName[table]; !ok {
		if len(b.tableNames) == 0 {
			// reserve the first alias for the main table
			b.tableNames = append(b.tableNames, sqls.Table{})
		}
		b.tableNames = append(b.tableNames, table)
	}
	tableAndAlias := table[0]
	if table[1] != "" {
		tableAndAlias = tableAndAlias + " AS " + table[1]
	}
	if on == nil || on.Raw == "" {
		b.tablesByName[table] = &fromTable{
			Table: table,
			Segment: &sqls.Segment{
				Raw: fmt.Sprintf("%s %s", joinStr, tableAndAlias),
			},
			Optional: optional,
		}
		return b
	}
	b.tablesByName[table] = &fromTable{
		Table: table,
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
