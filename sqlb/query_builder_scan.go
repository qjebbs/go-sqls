package sqlb

import (
	"database/sql"
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
)

// QueryScanner is the interface for QueryBuilder scanner, it tells
// QueryBuilder what to select and where to put to scanned values.
type QueryScanner interface {
	// Select tells QueryBuilder what to select, return nil to not do so, to
	// respect the columns set by *QueryBuilder.Select()
	Select() []*sqls.TableColumn
	// NewTarget creates a new scan target, it returns the target and its fields.
	NewTarget() (target any, fields []any)
}

// scanFunc is the function called when a row is scanned.
type scanFunc func(target any) error

// // ScanFunc scans query rows with scanner, and call the fn
// func (b *BaseQueryBuilder) ScanFunc(s Scanner, fn ScanFunc) error {
// 	_, err := b.scan(s, fn)
// 	return err
// }

// Scan scans query rows with scanner
func (b *QueryBuilder) Scan(s QueryScanner) ([]any, error) {
	return b.scan(s, nil)
}

// Scan scans query rows with scanner
func (b *QueryBuilder) scan(scanner QueryScanner, fn scanFunc) ([]any, error) {
	args := make([]any, 0)
	selects := scanner.Select()
	var (
		query string
		err   error
	)
	if len(selects) == 0 {
		query, err = b.buildInternal(&args, b.selects)
	} else {
		query, err = b.buildInternal(&args, &sqls.Segment{
			Prefix:  "SELECT",
			Raw:     "#join('#c', ', ')",
			Columns: selects,
		})
	}
	if err != nil {
		return nil, err
	}
	rows, err := b.db.Query(query, args...)
	if err != nil {
		query, _ := sqls.Interpolate(query, args...)
		return nil, fmt.Errorf("%w: %s", err, query)
	}
	defer rows.Close()

	var results []any
	bh := &blackhole{}
	for rows.Next() {
		target, fields := scanner.NewTarget()
		for i := 0; i < len(b.touches.Segments); i++ {
			fields = append(fields, &bh)
		}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}
		if fn == nil {
			results = append(results, target)
		} else {
			err = fn(target)
			if err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

type blackhole struct{}

func (b *blackhole) Scan(_ any) error { return nil }

// Count count the number of items that match the condition.
func (b *QueryBuilder) Count(columns ...*sqls.TableColumn) (count int64, err error) {
	args := make([]any, 0)
	query, err := b.buildInternal(&args, &sqls.Segment{
		Prefix:  "SELECT",
		Raw:     "#join('#c', ', ')",
		Columns: columns,
	})
	if err != nil {
		return 0, err
	}
	query = fmt.Sprintf(`SELECT COUNT(1) FROM (%s) list`, query)
	err = b.db.QueryRow(query, args...).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		query, _ := sqls.Interpolate(query, args...)
		return 0, fmt.Errorf("%w: %s", err, query)
	}
	return count, nil
}
