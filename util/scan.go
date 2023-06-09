package util

import (
	"database/sql"
	"fmt"

	"github.com/qjebbs/go-sqls"
)

// QueryAble is the interface for query-able *sql.DB, *sql.Tx
type QueryAble interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// NewScanDestFn is the function to create a new scan destination,
// returning the destination and the fields to scan.
type NewScanDestFn[T any] func() (T, []any)

// ScanBuilder is like Scan, but it builds query from sqls.Builder
func ScanBuilder[T any](db QueryAble, b sqls.Builder, fn NewScanDestFn[T]) ([]T, error) {
	query, args, err := b.Build()
	if err != nil {
		return nil, err
	}
	return Scan[T](db, query, args, fn)
}

// Scan scans query rows with scanner
func Scan[T any](db QueryAble, query string, args []any, fn NewScanDestFn[T]) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		query, _ := sqls.Interpolate(query, args...)
		return nil, fmt.Errorf("%w: %s", err, query)
	}
	defer rows.Close()

	var results []T
	bh := &blackhole{}
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		dest, fields := fn()
		nBlacholes := len(cols) - len(fields)
		for i := 0; i < nBlacholes; i++ {
			fields = append(fields, &bh)
		}
		err = rows.Scan(fields...)
		if err != nil {
			return nil, err
		}
		results = append(results, dest)
	}
	return results, nil
}

// CountBuilder is like Count, but it builds query from sqls.Builder
func CountBuilder(db QueryAble, b sqls.Builder) (count int64, err error) {
	query, args, err := b.Build()
	if err != nil {
		return 0, err
	}
	return Count(db, query, args)
}

// Count count the number of items that match the condition.
func Count(db QueryAble, query string, args []any) (count int64, err error) {
	query = fmt.Sprintf(`SELECT COUNT(1) FROM (%s) list`, query)
	err = db.QueryRow(query, args...).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		query, _ := sqls.Interpolate(query, args...)
		return 0, fmt.Errorf("%w: %s", err, query)
	}
	return count, nil
}

type blackhole struct{}

func (b *blackhole) Scan(_ any) error { return nil }
