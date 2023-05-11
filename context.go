package sqls

import (
	"fmt"
	"regexp"

	"github.com/qjebbs/go-sqls/syntax"
)

// GlobalContext is the global context shared between all segments building.
type GlobalContext struct {
	ArgStore     *[]any              // args store
	BindVarStyle syntax.BindVarStyle // bindvar style

	funcs FuncMap
}

// NewGlobalContext returns a new GlobalContext.
func NewGlobalContext(argStore *[]any) *GlobalContext {
	return &GlobalContext{
		ArgStore: argStore,
		funcs:    builtInFuncs(),
	}
}

// Funcs Funcs adds the elements of the argument map to the context's function map.
// It panics if a function name contains number [0-9].
func (ctx *GlobalContext) Funcs(funcs FuncMap) {
	r := regexp.MustCompile(`[0-9]`)
	for k, v := range funcs {
		if r.Match([]byte(k)) {
			panic(fmt.Errorf("function name %q contains number [0-9]", k))
		}
		ctx.funcs[k] = v
	}
}

// Context is the Context for current segment building.
type Context struct {
	Global  *GlobalContext // global context
	Segment *Segment       // current segment

	ArgsBuilt     []string // cache of built args
	ColumnsBuilt  []string // cache of built columns
	SegmentsBuilt []string // cache of built segments

	ArgsUsed     []bool // flags to indicate if an arg is used
	ColumnsUsed  []bool // flags to indicate if a column is used
	TableUsed    []bool // flag to indicate if a table is used
	SegmentsUsed []bool // flags to indicate if a segment is used
}

func newContext(ctx *GlobalContext, s *Segment) *Context {
	if s == nil {
		return nil
	}
	return &Context{
		Global:  ctx,
		Segment: s,

		ArgsBuilt:     make([]string, len(s.Args)),
		ColumnsBuilt:  make([]string, len(s.Columns)),
		SegmentsBuilt: make([]string, len(s.Segments)),

		ArgsUsed:     make([]bool, len(s.Args)),
		ColumnsUsed:  make([]bool, len(s.Columns)),
		TableUsed:    make([]bool, len(s.Tables)),
		SegmentsUsed: make([]bool, len(s.Segments)),
	}
}

func (c *Context) checkUsage() error {
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
