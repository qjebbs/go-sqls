package sqls

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

// TableColumn is a column of a table, it accpets a column Name or an
// Expression, the Name is prioritized if both set.
type TableColumn struct {
	Table Table
	Name  string

	// Expression accepts placeholders:
	// # for table alias; $1 or $1, $2 ... for args
	Expression string
	Args       []any
}

func (c *TableColumn) buildInternal(ctx *context) (string, error) {
	if c == nil {
		return "", nil
	}
	if c.Name != "" {
		switch {
		case c.Table[1] == "":
			return c.Name, nil
		default:
			return string(c.Table[1]) + "." + c.Name, nil
		}
	}
	seg := &Segment{
		Raw:     c.Expression,
		Args:    c.Args,
		Columns: c.Table.AnyColumns(),
	}
	clause, err := syntax.Parse(c.Expression)
	if err != nil || clause == nil {
		return "", err
	}
	ctx = ctx.newContextForSegment(seg)
	built, err := build(ctx, clause)
	if err != nil {
		return "", err
	}
	if err := ctx.checkArgUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", c.Expression, err)
	}
	return built, err
}

// Table holds the name and alias of a table or CTE (common table expression), e.g.:
//
//	users := Table{"users", "u"}
//	cte := Table{"cte", "c"}
type Table [2]string

func (t Table) String() string {
	if t[1] != "" {
		return t[1]
	}
	return t[0]
}

// AnyColumn returns a wildcard column of the table, e.g.:
//
//	t.AnyColumn() // "t.*"
func (t Table) AnyColumn() *TableColumn {
	return &TableColumn{
		Table: t,
		Name:  "*",
	}
}

// AnyColumns is the same as AnyColumn, but returns a slice.
func (t Table) AnyColumns() []*TableColumn {
	return []*TableColumn{
		t.AnyColumn(),
	}
}

// Column returns a named column of the table, e.g.:
//
//	t.Column("id") // "t.id"
func (t Table) Column(name string) *TableColumn {
	return &TableColumn{
		Table: t,
		Name:  name,
	}
}

// Expression returns a expression column of the table.
//
// The expression accepts placeholders:
//
//   - # => table alias
//   - $1, $2 ... => t.Args[0], t.Args[1] ...
//   - ?, ? ... => t.Args[0], t.Args[1] ...
//
// For example:
//
//	t.Expression("#t1.id")
//	t.Expression("COALESCE(#t1.id,0)")
//	t.Expression("#t1.deteled_at > $1", 1)
func (t Table) Expression(expression string, args ...any) *TableColumn {
	return &TableColumn{
		Table:      t,
		Expression: expression,
		Args:       args,
	}
}

// Columns returns the named columns of the table, e.g.:
//
//	t.Columns("id", "name")
func (t Table) Columns(names ...string) []*TableColumn {
	r := make([]*TableColumn, 0, len(names))
	for _, f := range names {
		r = append(r, &TableColumn{
			Table: t,
			Name:  f,
		})
	}
	return r
}

// Expressions returns expression columns of the table.
//
// The expressions accept placeholders:
//
//   - # => table alias
//   - $1, $2 ... => t.Args[0], t.Args[1] ...
//   - ?, ? ... => t.Args[0], t.Args[1] ...
//
// For example:
//
//	t.Expressions("#t1.id", "#t1.deteled_at")
//	t.Expressions("COALESCE(#t1.id,0)", "#t1.deteled_at IS NULL")
func (t Table) Expressions(expressions ...string) []*TableColumn {
	r := make([]*TableColumn, 0, len(expressions))
	for _, exp := range expressions {
		r = append(r, &TableColumn{
			Table:      t,
			Expression: exp,
		})
	}
	return r
}
