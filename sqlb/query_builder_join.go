package sqlb

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

// From set the from table.
func (b *QueryBuilder) From(t Table) *QueryBuilder {
	if t.Name == "" {
		b.pushError(fmt.Errorf("from table is empty"))
		return b
	}
	tableAndAlias := string(t.Name)
	if t.Alias != "" {
		tableAndAlias = tableAndAlias + " AS " + string(t.Alias)
	}
	if len(b.tables) == 0 {
		b.tables = append(b.tables, t)
	} else {
		b.tables[0] = t
	}
	b.appliedNames[t.AppliedName()] = t
	b.froms[t] = &fromTable{
		Segment: &sqls.Segment{
			Raw: tableAndAlias,
		},
		Optional: false,
	}
	return b
}

// InnerJoin append a inner join table.
func (b *QueryBuilder) InnerJoin(t Table, on *sqls.Segment) *QueryBuilder {
	return b.join("INNER JOIN", t, on, false)
}

// LeftJoin append / replace a left join table.
func (b *QueryBuilder) LeftJoin(t Table, on *sqls.Segment) *QueryBuilder {
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
func (b *QueryBuilder) LeftJoinOptional(t Table, on *sqls.Segment) *QueryBuilder {
	return b.join("LEFT JOIN", t, on, true)
}

// RightJoin append / replace a right join table.
func (b *QueryBuilder) RightJoin(t Table, on *sqls.Segment) *QueryBuilder {
	return b.join("RIGHT JOIN", t, on, false)
}

// FullJoin append / replace a full join table.
func (b *QueryBuilder) FullJoin(t Table, on *sqls.Segment) *QueryBuilder {
	return b.join("FULL JOIN", t, on, false)
}

// CrossJoin append / replace a cross join table.
func (b *QueryBuilder) CrossJoin(t Table) *QueryBuilder {
	return b.join("CROSS JOIN", t, nil, false)
}

// From append or replace a from table.
func (b *QueryBuilder) join(joinStr string, t Table, on *sqls.Segment, optional bool) *QueryBuilder {
	if t.Name == "" {
		b.pushError(fmt.Errorf("join table name is empty"))
		return b
	}
	if _, ok := b.froms[t]; ok {
		if t.Alias == "" {
			b.pushError(fmt.Errorf("table [%s] is already joined", t.Name))
			return b
		}
		b.pushError(fmt.Errorf("table [%s AS %s] is already joined", t.Name, t.Alias))
		return b
	}
	if len(b.tables) == 0 {
		// reserve the first alias for the main table
		b.tables = append(b.tables, Table{})
	}
	b.tables = append(b.tables, t)
	b.appliedNames[t.AppliedName()] = t
	tableAndAlias := t.Name
	if t.Alias != "" {
		tableAndAlias = tableAndAlias + " AS " + t.Alias
	}
	if on == nil || on.Raw == "" {
		b.froms[t] = &fromTable{
			Segment: &sqls.Segment{
				Raw: fmt.Sprintf("%s %s", joinStr, tableAndAlias),
			},
			Optional: optional,
		}
		return b
	}
	b.froms[t] = &fromTable{
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
