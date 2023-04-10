package sqls

import (
	"fmt"

	"github.com/qjebbs/go-sqls/syntax"
)

// Context is the global context shared between all segments building.
type Context struct {
	ArgStore     *[]any              // args store
	BindVarStyle syntax.BindVarStyle // bindvar style

	funcMap map[string]preprocessor // func map
}

// NewContext returns a new context.
func NewContext(argStore *[]any) *Context {
	return &Context{
		ArgStore: argStore,
		funcMap:  builtInFuncs,
	}
}

// context is the context for current segment building.
type context struct {
	global  *Context // global context
	Segment *Segment // current segment

	ArgsBuilt     []string // cache of built args
	ColumnsBuilt  []string // cache of built columns
	SegmentsBuilt []string // cache of built segments

	ArgsUsed     []bool // flags to indicate if an arg is used
	ColumnsUsed  []bool // flags to indicate if a column is used
	TableUsed    []bool // flag to indicate if a table is used
	SegmentsUsed []bool // flags to indicate if a segment is used
}

func newSegmentContext(ctx *Context, s *Segment) *context {
	if s == nil {
		return nil
	}
	return &context{
		global:        ctx,
		Segment:       s,
		ArgsBuilt:     make([]string, len(s.Args)),
		ColumnsBuilt:  make([]string, len(s.Columns)),
		TableUsed:     make([]bool, len(s.Tables)),
		SegmentsBuilt: make([]string, len(s.Segments)),
		ArgsUsed:      make([]bool, len(s.Args)),
		ColumnsUsed:   make([]bool, len(s.Columns)),
		SegmentsUsed:  make([]bool, len(s.Segments)),
	}
}

func (c *context) checkUsage() error {
	if c == nil {
		return nil
	}
	for i, v := range c.ArgsUsed {
		if !v {
			return fmt.Errorf("arg %d is not used", i+1)
		}
	}
	for i, v := range c.ColumnsUsed {
		if !v {
			return fmt.Errorf("column %d is not used", i+1)
		}
	}
	for i, v := range c.TableUsed {
		if !v {
			return fmt.Errorf("table %d is not used", i+1)
		}
	}
	for i, v := range c.SegmentsUsed {
		if !v {
			return fmt.Errorf("segment %d is not used", i+1)
		}
	}
	return nil
}
