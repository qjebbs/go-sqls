package sqlb

import (
	"database/sql"
	"fmt"

	"github.com/qjebbs/go-sqls"
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

// Scan scans query rows with scanner
func (b *QueryBuilder) Scan(s QueryScanner) ([]any, error) {
	args := make([]any, 0)
	ctx := sqls.NewContext(&args)
	ctx.BindVarStyle = b.bindVarStyle
	selects := s.Select()
	var (
		query string
		err   error
	)
	if len(selects) == 0 {
		query, err = b.buildInternal(ctx, b.selects)
	} else {
		query, err = b.buildInternal(ctx, &sqls.Segment{
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
		target, fields := s.NewTarget()
		for i := 0; i < len(b.touches.Segments); i++ {
			fields = append(fields, &bh)
		}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}
		results = append(results, target)
	}
	return results, nil
}

type blackhole struct{}

func (b *blackhole) Scan(_ any) error { return nil }

// Count count the number of items that match the condition.
func (b *QueryBuilder) Count(columns ...*sqls.TableColumn) (count int64, err error) {
	args := make([]any, 0)
	ctx := sqls.NewContext(&args)
	ctx.BindVarStyle = b.bindVarStyle
	query, err := b.buildInternal(ctx, &sqls.Segment{
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
