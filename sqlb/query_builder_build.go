package sqlb

import (
	"fmt"
	"strings"

	"git.qjebbs.com/jebbs/go-sqls"
)

// Build builds the query.
func (b *QueryBuilder) Build() (query string, args []any, err error) {
	args = make([]any, 0)
	query, err = b.buildInternal(&args, b.selects)
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

// BuildTo builds the query to the argStore.
func (b *QueryBuilder) BuildTo(argStore *[]any) (query string, err error) {
	return b.buildInternal(argStore, b.selects)
}

// buildInternal builds the query with the selects.
func (b *QueryBuilder) buildInternal(argStore *[]any, selects *sqls.Segment) (string, error) {
	if b == nil {
		return "", nil
	}
	if err := b.anyError(); err != nil {
		return "", err
	}
	clauses := make([]string, 0)

	dep, err := b.calcDependency(selects)
	if err != nil {
		return "", err
	}

	sq, err := b.buildCTEs(argStore, dep)
	if err != nil {
		return "", err
	}
	if sq != "" {
		clauses = append(clauses, sq)
	}

	sel, err := b.buildSelects(argStore, selects)
	if err != nil {
		return "", err
	}
	clauses = append(clauses, sel)
	from, err := b.buildFrom(argStore, dep)
	if err != nil {
		return "", err
	}
	if from != "" {
		clauses = append(clauses, from)
	}
	where, err := b.conditions.BuildTo(argStore)
	if err != nil {
		return "", err
	}
	if where != "" {
		clauses = append(clauses, where)
	}
	groupby, err := b.groupbys.BuildTo(argStore)
	if err != nil {
		return "", err
	}
	if groupby != "" {
		clauses = append(clauses, groupby)
	}
	order, err := b.orders.BuildTo(argStore)
	if err != nil {
		return "", err
	}
	if order != "" {
		clauses = append(clauses, order)
	}
	if b.limit > 0 {
		clauses = append(clauses, fmt.Sprintf(`LIMIT %d`, b.limit))
	}
	if b.offset > 0 {
		clauses = append(clauses, fmt.Sprintf(`OFFSET %d`, b.offset))
	}
	query := strings.Join(clauses, " ")
	if len(b.unions) == 0 {
		return strings.TrimSpace(query), nil
	}
	union, err := b.buildUnion(argStore)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(query + " " + union), nil
}

func (b *QueryBuilder) buildCTEs(argStore *[]any, dep map[sqls.Table]bool) (string, error) {
	if len(b.commonTableExprs) == 0 {
		return "", nil
	}
	clauses := make([]string, 0, len(b.commonTableExprs))
	for _, cte := range b.commonTableExprs {
		table, ok := b.tablesByName[cte.alias]
		if !ok {
			// the join clause refers to the CTE may be replaced
			continue
		}
		if b.distinct && table.Optional && !dep[cte.alias] {
			continue
		}
		query, err := cte.BuildTo(argStore)
		if err != nil {
			return "", fmt.Errorf("build subquery '%s (%s)': %w", cte.name, cte.alias, err)
		}
		if query == "" {
			continue
		}
		clauses = append(clauses, fmt.Sprintf(
			"%s AS (%s)",
			cte.name, query,
		))
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "With " + strings.Join(clauses, ", "), nil
}

func (b *QueryBuilder) buildSelects(argStore *[]any, s *sqls.Segment) (string, error) {
	if b.distinct {
		s.Prefix = "SELECT DISTINCT"
	} else {
		s.Prefix = "SELECT"
	}
	sel, err := s.BuildTo(argStore)
	if err != nil {
		return "", err
	}
	touches, err := b.touches.BuildTo(argStore)
	if err != nil {
		return "", err
	}
	if sel == "" {
		return "", fmt.Errorf("no columns selected")
	}
	if touches == "" {
		return sel, nil
	}
	return sel + ", " + touches, nil
}

func (b *QueryBuilder) buildFrom(argStore *[]any, dep map[sqls.Table]bool) (string, error) {
	tables := make([]string, 0, len(b.tableNames))
	for _, t := range b.tableNames {
		ft, ok := b.tablesByName[t]
		if !ok {
			// should not happen
			return "", fmt.Errorf("table '%s' not found", t)
		}
		if b.distinct && ft.Optional && !dep[t] {
			continue
		}
		from := b.tablesByName[t]
		c, err := from.Segment.BuildTo(argStore)
		if err != nil {
			return "", fmt.Errorf("build FROM '%s': %w", from.Segment.Raw, err)
		}
		tables = append(tables, c)
	}
	return "FROM " + strings.Join(tables, " "), nil
}

func (b *QueryBuilder) buildUnion(argStore *[]any) (string, error) {
	clauses := make([]string, 0, len(b.unions))
	for _, union := range b.unions {
		query, err := union.BuildTo(argStore)
		if err != nil {
			return "", err
		}
		if query == "" {
			continue
		}
		clauses = append(clauses, query)
	}
	return "UNION (" + strings.Join(clauses, ") UNION (") + ")", nil
}
