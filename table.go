package sqls

// Table is a table identifier, it can a table name or an alias.
type Table string

// Column returns a column of the table.
// It add table prefix to the column name, e.g.: "id" -> "t.id".
//
// For example:
//
//	t := Table("t")
//	// these two are equivalent
//	t.Column("id")         // "t.id"
//	t.Expression("#t1.id") // "t.id"
//
// If you want to use the column name directly, try:
//
//	t.Expressions("id") // "id"
func (t Table) Column(name string) *TableColumn {
	return &TableColumn{
		Table: t,
		Raw:   "#t1." + name,
	}
}

// Columns returns columns of the table from names.
// It add table prefix to the column name, e.g.: "id" -> "t.id".
//
// For example:
//
//	t := Table("t")
//	// these two are equivalent
//	t.Columns("id", "name")              // "t.id", "t.name"
//	t.Expressions("#t1.id", "#t1.name")  // "t.id", "t.name"
//
// If you want to use the column name directly, try:
//
//	t.Expressions("id", "name") // "id", "name"
func (t Table) Columns(names ...string) []*TableColumn {
	r := make([]*TableColumn, 0, len(names))
	for _, name := range names {
		r = append(r, t.Column(name))
	}
	return r
}

// Expression returns a column of the table from the expression which
// accepts bindvars and the preprocessor #t1 (name), #t1  (alias), which
// are implicit in "*TableColumn.Table".
//
// For example:
//
//	t := Table("t")
//	t.Expression("id")                       // "id"
//	t.Expression("#t1.id")                   // "table.id"
//	t.Expression("#t1.id")                  // "t.id"
//	t.Expression("COALESCE(#t1.id,0)")      // "COALESCE(t.id,0)"
//	t.Expression("#t1.deteled_at > $1", 1)  // "t.deteled_at > $1"
func (t Table) Expression(expression string, args ...any) *TableColumn {
	return &TableColumn{
		Table: t,
		Raw:   expression,
		Args:  args,
	}
}

// Expressions returns columns of the table from the expression, which
// accepts bindvars and the preprocessor #t1 (name), #t1  (alias) which
// are implicit in "*TableColumn.Table".
//
// For example:
//
//	t := Table("t")
//	t.Expressions("id", "deteled_at")         // "id", "deteled_at"
//	t.Expressions("#t1.id", "#t1.deteled_at") // "table.id", "table.deteled_at"
//	t.Expressions("#t1.id", "#t1.deteled_at") // "t.id", "t.deteled_at"
//	t.Expressions("COALESCE(#t1.id,0)")       // "COALESCE(t.id,0)"
func (t Table) Expressions(expressions ...string) []*TableColumn {
	r := make([]*TableColumn, 0, len(expressions))
	for _, exp := range expressions {
		r = append(r, t.Expression(exp))
	}
	return r
}

// TableColumn is a column of a table.
type TableColumn struct {
	Table Table
	Raw   string
	Args  []any
}
