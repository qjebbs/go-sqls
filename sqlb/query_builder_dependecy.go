package sqlb

import (
	"fmt"

	"github.com/qjebbs/go-sqls"
	"github.com/qjebbs/go-sqls/slices"
)

func (b *QueryBuilder) calcDependency(selects *sqls.Segment) (map[Table]bool, error) {
	columns := slices.Concat(
		extractColumns(selects),
		extractColumns(b.touches),
		extractColumns(b.conditions),
		extractColumns(b.orders),
		extractColumns(b.groupbys),
	)
	m := make(map[Table]bool)
	// first table is the main table and always included
	m[b.tables[0]] = true
	for _, column := range columns {
		err := b.markDependencies(m, column.Table)
		if err != nil {
			return nil, err
		}
	}
	// mark for CTEs
	for _, t := range b.tables {
		if b.distinct && b.froms[t].Optional && !m[t] {
			continue
		}
		// this could probably mark a CTE table that does not exists, but do no harm.
		m[NewTable(t.Name, "")] = true
	}
	return m, nil
}

func (b *QueryBuilder) markDependencies(dep map[Table]bool, t sqls.Table) error {
	ta, ok := b.appliedNames[t]
	if !ok {
		return fmt.Errorf("table not found: '%s'", t)
	}
	from, ok := b.froms[ta]
	if !ok {
		return fmt.Errorf("from undefined: '%s'", t)
	}
	if dep[ta] {
		return nil
	}
	dep[ta] = true
	for _, column := range from.Segment.Columns {
		if column.Table == t {
			continue
		}
		err := b.markDependencies(dep, t)
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
