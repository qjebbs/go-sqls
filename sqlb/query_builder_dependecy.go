package sqlb

import (
	"fmt"

	"github.com/qjebbs/go-sqls"
)

func (b *QueryBuilder) calcDependency(selects *sqls.Segment) (map[Table]bool, error) {
	tables := extractTables(
		selects,
		b.touches,
		b.conditions,
		b.orders,
		b.groupbys,
	)
	m := make(map[Table]bool)
	// first table is the main table and always included
	m[b.tables[0]] = true
	for _, t := range tables {
		err := b.markDependencies(m, t.Table)
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
	for _, ft := range extractTables(from.Segment) {
		if ft.Table == t {
			continue
		}
		err := b.markDependencies(dep, ft.Table)
		if err != nil {
			return fmt.Errorf("%s: %s", ft.Source, err)
		}
	}
	return nil
}

type tableWithSouce struct {
	Table  sqls.Table
	Source string
}

func extractTables(segments ...*sqls.Segment) []*tableWithSouce {
	tables := []*tableWithSouce{}
	dict := map[sqls.Table]bool{}
	extractTables2(segments, &tables, &dict)
	return tables
}

func extractTables2(segments []*sqls.Segment, tables *[]*tableWithSouce, dict *map[sqls.Table]bool) {
	for _, s := range segments {
		if s == nil {
			continue
		}
		for i, t := range s.Tables {
			if (*dict)[t] {
				continue
			}
			*tables = append(*tables, &tableWithSouce{
				Table:  t,
				Source: fmt.Sprintf("#tables%d of '%s'", i+1, s.Raw),
			})
			(*dict)[t] = true
		}
		for i, c := range s.Columns {
			if c == nil || (*dict)[c.Table] {
				continue
			}
			*tables = append(*tables, &tableWithSouce{
				Table:  c.Table,
				Source: fmt.Sprintf("#column%d '%s' of '%s'", i+1, c.Raw, s.Raw),
			})
			(*dict)[c.Table] = true
		}
		extractTables2(s.Segments, tables, dict)
	}
}
