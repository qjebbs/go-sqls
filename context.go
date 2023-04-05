package sqls

import (
	"fmt"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

// context is the context for building.
type context struct {
	Global  *contextGlobal
	Current *contextCurrent
}

// newContext returns a new context.
func newContext(argStore *[]any) *context {
	return &context{
		Global: &contextGlobal{
			ArgStore: argStore,
			FuncMap:  builtInFuncs,
		},
	}
}

// WithSegment returns a new context with the given segment.
func (c *context) WithSegment(s *Segment) *context {
	if s == nil {
		return nil
	}
	return &context{
		Global: c.Global,
		Current: &contextCurrent{
			Segment:       s,
			ArgsBuilt:     make([]string, len(s.Args)),
			ColumnsBuilt:  make([]string, len(s.Columns)),
			TableUsed:     make([]bool, len(s.Tables)),
			SegmentsBuilt: make([]string, len(s.Segments)),
			ArgsUsed:      make([]bool, len(s.Args)),
			ColumnsUsed:   make([]bool, len(s.Columns)),
			SegmentsUsed:  make([]bool, len(s.Segments)),
		},
	}
}

// contextGlobal is the global context shared between all segments building.
type contextGlobal struct {
	ArgStore *[]any                  // args store
	FuncMap  map[string]preprocessor // func map

	BindVarStyle syntax.RefType // bindvar style
	FirstBindvar string         // first bindvar seen
}

// contextCurrent is the context for current segment building.
type contextCurrent struct {
	Segment *Segment // current segment

	ArgsBuilt     []string // cache of built args
	ColumnsBuilt  []string // cache of built columns
	SegmentsBuilt []string // cache of built segments

	ArgsUsed     []bool // flags to indicate if an arg is used
	ColumnsUsed  []bool // flags to indicate if a column is used
	TableUsed    []bool // flag to indicate if a table is used
	SegmentsUsed []bool // flags to indicate if a segment is used
}

func (c *contextCurrent) checkUsage() error {
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
