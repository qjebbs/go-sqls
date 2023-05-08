package sqlb

import (
	"database/sql"
	"fmt"

	"github.com/qjebbs/go-sqls"
	"github.com/qjebbs/go-sqls/syntax"
)

var _ sqls.Builder = (*QueryBuilder)(nil)

// QueryBuilder is the SQL query builder.
// It's recommended to wrap it with your struct to provide a
// more friendly API and improve segment reusability.
type QueryBuilder struct {
	db           QueryAble           // the database connection
	bindVarStyle syntax.BindVarStyle // the bindvar style

	ctes         []*cte               // common table expressions
	froms        map[Table]*fromTable // the from tables by alias
	tables       []Table              // the tables in order
	appliedNames map[sqls.Table]Table // applied table name mapping, the name is alias, or name if alias is empty

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
	Segment  *sqls.Segment
	Optional bool
}

// NewQueryBuilder returns a new QueryBuilder.
func NewQueryBuilder(db QueryAble) *QueryBuilder {
	return &QueryBuilder{
		db:           db,
		froms:        map[Table]*fromTable{},
		appliedNames: make(map[sqls.Table]Table),
		selects: &sqls.Segment{
			Prefix: "SELECT",
			Raw:    "#join('#column', ', ')",
		},
		touches: &sqls.Segment{
			Prefix: "",
			Raw:    "#join('#segment', ', ')",
		},
		conditions: &sqls.Segment{
			Prefix: "WHERE",
			Raw:    "#join('#segment', ' AND ')",
		},
		orders: &sqls.Segment{
			Prefix: "ORDER BY",
			Raw:    "#join('#segment', ', ')",
		},
		groupbys: &sqls.Segment{
			Prefix: "GROUP BY",
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

// Order is the sorting order.
type Order uint

// orders
const (
	Asc Order = iota
	AscNullsFirst
	AscNullsLast
	Desc
	DescNullsFirst
	DescNullsLast
)

var orders = []string{
	"ASC",
	"ASC NULLS FIRST",
	"ASC NULLS LAST",
	"DESC",
	"DESC NULLS FIRST",
	"DESC NULLS LAST",
}

// OrderBy set the sorting order. the order can be "ASC", "DESC", "ASC NULLS FIRST" or "DESC NULLS LAST"
func (b *QueryBuilder) OrderBy(column *sqls.TableColumn, order Order) *QueryBuilder {
	idx := len(b.orders.Segments) + 1
	alias := fmt.Sprintf("_order_%d", idx)

	if order > DescNullsLast {
		b.pushError(fmt.Errorf("invalid order: %d", order))
	}
	orderStr := orders[order]
	// pq: for SELECT DISTINCT, ORDER BY expressions must appear in select list
	b.touches.AppendSegments(&sqls.Segment{
		Raw:     "#c1 AS " + alias,
		Columns: []*sqls.TableColumn{column},
	})
	b.orders.AppendSegments(&sqls.Segment{
		Raw:     fmt.Sprintf("%s %s", alias, orderStr),
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

// BindVar set the bindvar style.
func (b *QueryBuilder) BindVar(style syntax.BindVarStyle) *QueryBuilder {
	b.bindVarStyle = style
	return b
}
