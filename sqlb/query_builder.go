package sqlb

import (
	"database/sql"
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

var _ sqls.Builder = (*QueryBuilder)(nil)

// QueryBuilder is the SQL query builder.
// It's recommended to wrap it with your struct to provide a
// more friendly API and improve segment reusability.
type QueryBuilder struct {
	db QueryAble // the database connection

	commonTableExprs []*namedSegment           // common table expressions
	tableNames       []sqls.Table              // the table names
	tablesByName     map[sqls.Table]*fromTable // the tables by name

	selects    *sqls.Segment  // select columns and keep values in scanning.
	touches    *sqls.Segment  // select columns but drop values in scanning.
	conditions *sqls.Segment  // where conditions, joined with AND.
	orders     *sqls.Segment  // order by columns, joined with comma.
	groupbys   *sqls.Segment  // group by columns, joined with comma.
	distinct   bool           // select distinct
	limit      int64          // limit count
	offset     int64          // offset count
	unions     []sqls.Builder // union queries

	errors []error // errors during building
}

// QueryAble is the interface for query-able *sql.DB, *sql.Tx
type QueryAble interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type fromTable struct {
	Table    sqls.Table
	Segment  *sqls.Segment
	Optional bool
}

// NewQueryBuilder returns a new QueryBuilder.
func NewQueryBuilder(db QueryAble) *QueryBuilder {
	return &QueryBuilder{
		db:           db,
		tablesByName: make(map[sqls.Table]*fromTable),
		selects: &sqls.Segment{
			Header: "SELECT",
			Raw:    "#join('#column', ', ')",
		},
		touches: &sqls.Segment{
			Header: "",
			Raw:    "#join('#segment', ', ')",
		},
		conditions: &sqls.Segment{
			Header: "WHERE",
			Raw:    "#join('#segment', ' AND ')",
		},
		orders: &sqls.Segment{
			Header: "ORDER BY",
			Raw:    "#join('#segment', ', ')",
		},
		groupbys: &sqls.Segment{
			Header: "GROUP BY",
			Raw:    "#join('#segment', ', ')",
		},
	}
}

// Distinct set the flag for SELECT DISTINCT.
func (b *QueryBuilder) Distinct() *QueryBuilder {
	b.distinct = true
	return b
}

// Select replace the SELECT clause with the columns.
func (b *QueryBuilder) Select(columns []*sqls.TableColumn) *QueryBuilder {
	if len(columns) == 0 {
		return b
	}
	b.selects.AppendColumns(columns...)
	return b
}

// OrderBy set the sorting order. the order can be "ASC", "DESC", "ASC NULLS FIRST" or "DESC NULLS LAST"
func (b *QueryBuilder) OrderBy(column *sqls.TableColumn, order string, args ...any) *QueryBuilder {
	idx := len(b.orders.Segments) + 1
	alias := fmt.Sprintf("_order_%d", idx)

	// pq: for SELECT DISTINCT, ORDER BY expressions must appear in select list
	b.touches.AppendSegments(&sqls.Segment{
		Raw:     "#c1 AS " + alias,
		Columns: []*sqls.TableColumn{column},
		Args:    args,
	})
	b.orders.AppendSegments(&sqls.Segment{
		Raw:     fmt.Sprintf("%s %s", alias, order),
		Columns: nil,
		Args:    nil,
	})
	return b
}

// Limit set the limit.
func (b *QueryBuilder) Limit(limit int64) *QueryBuilder {
	if limit > 0 {
		b.limit = limit
	}
	return b
}

// Offset set the offset.
func (b *QueryBuilder) Offset(offset int64) *QueryBuilder {
	if offset > 0 {
		b.offset = offset
	}
	return b
}

// GroupBy set the sorting order.
func (b *QueryBuilder) GroupBy(column *sqls.TableColumn, args ...any) *QueryBuilder {
	b.groupbys.AppendSegments(&sqls.Segment{
		Raw:     "#c1",
		Columns: []*sqls.TableColumn{column},
		Args:    args,
	})
	return b
}

// Union unions other query builders, the type of query builders can be
// *QueryBuilder or any other extended *QueryBuilder types (structs with
// *QueryBuilder embedded.)
func (b *QueryBuilder) Union(builders ...sqls.Builder) *QueryBuilder {
	b.unions = append(b.unions, builders...)
	return b
}

// From set the from table.
func (b *QueryBuilder) From(t sqls.Table) *QueryBuilder {
	if t[1] == "" {
		b.pushError(fmt.Errorf("join table alias is empty"))
		return b
	}
	tableAndAlias := t[1]
	if t[0] != "" {
		tableAndAlias = t[0] + " " + tableAndAlias
	}
	if len(b.tableNames) == 0 {
		b.tableNames = append(b.tableNames, t)
	} else {
		b.tableNames[0] = t
	}
	b.tablesByName[t] = &fromTable{
		Table: t,
		Segment: &sqls.Segment{
			Raw: tableAndAlias,
		},
		Optional: false,
	}
	return b
}
