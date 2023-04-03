package sqlb

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls"
	"git.qjebbs.com/jebbs/go-sqls/slices"
)

func (b *QueryBuilder) calcDependency(selects *sqls.Segment) (map[sqls.Table]bool, error) {
	columns := slices.Concat(
		extractColumns(selects),
		extractColumns(b.touches),
		extractColumns(b.conditions),
		extractColumns(b.orders),
		extractColumns(b.groupbys),
	)
	m := make(map[sqls.Table]bool)
	// first table is the main table and always included
	m[b.tableNames[0]] = true
	for _, column := range columns {
		err := b.markDependencies(m, column.Table)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (b *QueryBuilder) markDependencies(m map[sqls.Table]bool, t sqls.Table) error {
	from, ok := b.tablesByName[t]
	if !ok {
		return fmt.Errorf("from undefined: '%s'", t)
	}
	if m[t] {
		return nil
	}
	m[t] = true
	for _, column := range from.Segment.Columns {
		if column.Table == t {
			continue
		}
		err := b.markDependencies(m, column.Table)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractColumns(s ...*sqls.Segment) []*sqls.TableColumn {
	f := make([]*sqls.TableColumn, 0, len(s))
	for _, s := range s {
		if s == nil {
			continue
		}
		f = append(f, s.Columns...)
		for _, s := range s.Segments {
			f = append(f, extractColumns(s)...)
		}
	}
	return f
}
