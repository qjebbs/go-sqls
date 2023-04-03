package sqls

import (
	"fmt"
	"strconv"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

type context struct {
	*globalContext
	*perSegmentContext
}

// Func is the type of function that can be registered to context.
type Func func(ctx *context, args ...string) (string, error)

func newContext(argStore *[]any, s *Segment) *context {
	return (&context{
		globalContext: &globalContext{
			argStore: argStore,
			funcMap: map[string]Func{
				"join":    join,
				"$":       argumentDollar,
				"?":       argumentQuestion,
				"c":       column,
				"col":     column,
				"column":  column,
				"t":       table,
				"table":   table,
				"tAs":     tableNameAndAlias,
				"s":       segment,
				"seg":     segment,
				"segment": segment,
			},
		},
	}).newContextForSegment(s)
}

func (c *context) newContextForSegment(s *Segment) *context {
	if s == nil {
		return c
	}
	return &context{
		globalContext: c.globalContext,
		perSegmentContext: &perSegmentContext{
			s:             s,
			argsCache:     make([]string, len(s.Args)),
			columnsCache:  make([]string, len(s.Columns)),
			segmentsCache: make([]string, len(s.Segments)),
			columnsUsed:   make([]bool, len(s.Columns)),
			segmentsUsed:  make([]bool, len(s.Segments)),
		},
	}
}

func (c *context) newContextForColumn(col *TableColumn) *context {
	if col == nil {
		return c
	}
	return &context{
		globalContext: c.globalContext,
		perSegmentContext: &perSegmentContext{
			s:         &Segment{Args: col.Args},
			argsCache: make([]string, len(col.Args)),
		},
	}
}

type globalContext struct {
	argStore *[]any
	funcMap  map[string]Func

	bindVarStyle    syntax.RefType
	firstBindvar    string
	firstBindvarPos syntax.Pos
}

type perSegmentContext struct {
	s *Segment

	argsCache     []string
	columnsCache  []string
	segmentsCache []string

	columnsUsed  []bool
	segmentsUsed []bool
}

func (c *perSegmentContext) Arg(ctx *context, index int) (string, error) {
	if index > len(c.s.Args) {
		return "", fmt.Errorf("invalid bindvar index %d", index)
	}
	i := index - 1
	built := c.argsCache[i]
	if built == "" || ctx.bindVarStyle == syntax.ArgUnindexed {
		*ctx.argStore = append(*ctx.argStore, c.s.Args[i])
		if ctx.bindVarStyle == syntax.ArgUnindexed {
			built = "?"
		} else {
			built = "$" + strconv.Itoa(len(*ctx.argStore))
		}
		c.argsCache[i] = built
	}
	return built, nil
}

func (c *perSegmentContext) Column(ctx *context, index int) (string, error) {
	if index > len(c.s.Columns) {
		return "", fmt.Errorf("invalid column index %d", index)
	}
	i := index - 1
	col := c.s.Columns[i]
	built := c.columnsCache[i]
	if built == "" || (ctx.bindVarStyle == syntax.ArgUnindexed && len(col.Args) > 0) {
		b, err := col.buildInternal(ctx.newContextForColumn(col))
		if err != nil {
			return "", err
		}
		c.columnsCache[i] = b
		c.columnsUsed[i] = true
		built = b
	}
	return built, nil
}

func (c *perSegmentContext) Table(ctx *context, index int) (string, error) {
	c.columnsUsed[index-1] = true
	if index > len(c.s.Columns) {
		return "", fmt.Errorf("invalid table index %d", index)
	}
	return c.s.Columns[index-1].Table.String(), nil
}

func (c *perSegmentContext) TableAndAaias(ctx *context, index int) (string, error) {
	c.columnsUsed[index-1] = true
	if index > len(c.s.Columns) {
		return "", fmt.Errorf("invalid table index %d", index)
	}
	t := c.s.Columns[index-1].Table
	if t[0] == "" {
		return "", fmt.Errorf("empty table name at %d", index)
	}
	if t[1] == "" {
		return t[0], nil
	}
	return t[0] + " AS " + t[1], nil
}

func (c *perSegmentContext) Segment(ctx *context, index int) (string, error) {
	if index > len(c.s.Segments) {
		return "", fmt.Errorf("invalid segment index %d", index)
	}
	i := index - 1
	seg := c.s.Segments[i]
	built := c.segmentsCache[i]
	if built == "" || (ctx.bindVarStyle == syntax.ArgUnindexed && len(seg.Args) > 0) {
		b, err := seg.buildInternal(ctx.newContextForSegment(seg))
		if err != nil {
			return "", err
		}
		c.segmentsCache[i] = b
		c.segmentsUsed[i] = true
		built = b
	}
	return built, nil
}

func (c *perSegmentContext) checkUsage() error {
	for i, v := range c.argsCache {
		if v == "" {
			return fmt.Errorf("arg %d is not used", i+1)
		}
	}
	for i, v := range c.columnsUsed {
		if !v {
			return fmt.Errorf("column %d is not used", i+1)
		}
	}
	for i, v := range c.segmentsUsed {
		if !v {
			return fmt.Errorf("segment %d is not used", i+1)
		}
	}
	return nil
}

func (c *perSegmentContext) checkArgUsage() error {
	for i, v := range c.argsCache {
		if v == "" {
			return fmt.Errorf("arg %d is not used", i+1)
		}
	}
	return nil
}
