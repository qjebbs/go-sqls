package sqls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qjebbs/go-sqls/syntax"
)

// Build builds the segment.
func (s *Segment) Build() (query string, args []any, err error) {
	args = make([]any, 0)
	query, err = s.BuildContext(NewContext(&args))
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

// BuildContext builds the segment with context.
func (s *Segment) BuildContext(ctx *Context) (string, error) {
	if s == nil {
		return "", nil
	}
	ctxCur := newSegmentContext(ctx, s)
	body, err := build(ctxCur)
	if err != nil {
		return "", err
	}
	if err := ctxCur.checkUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", ctxCur.Segment.Raw, err)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return "", nil
	}
	header, footer := "", ""
	if s.Prefix != "" {
		header = s.Prefix + " "
	}
	if s.Suffix != "" {
		footer = " " + s.Suffix
	}
	return header + body + footer, nil
}

// build builds the segment
func build(ctx *context) (string, error) {
	clause, err := syntax.Parse(ctx.Segment.Raw)
	if err != nil {
		return "", fmt.Errorf("parse '%s': %w", ctx.Segment.Raw, err)
	}
	built, err := buildCluase(ctx, clause)
	if err != nil {
		return "", fmt.Errorf("build '%s': %w", ctx.Segment.Raw, err)
	}
	return built, nil
}

// buildCluase builds the parsed clause within current context, not updating the ctx.current.
func buildCluase(ctx *context, clause *syntax.Clause) (string, error) {
	b := new(strings.Builder)
	for _, decl := range clause.ExprList {
		switch expr := decl.(type) {
		case *syntax.PlainExpr:
			b.WriteString(expr.Text)
		case *syntax.FuncCallExpr:
			fn := builtInFuncs[expr.Name]
			if fn == nil {
				return "", fmt.Errorf("function '%s' is not found", expr.Name)
			}
			s, err := fn(ctx, expr.Args...)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		case *syntax.BindVarExpr:
			if ctx.global.BindVarStyle == 0 {
				ctx.global.BindVarStyle = expr.Type
				// ctx.global.FirstBindvar = ctx.Segment.Raw
			}
			// if ctx.global.BindVarStyle != expr.Type {
			// 	return "", fmt.Errorf("mixed bindvar styles between segments '%s' and '%s'", ctx.global.FirstBindvar, ctx.Segment.Raw)
			// }
			s, err := buildArg(ctx, expr.Index)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		default:
			return "", fmt.Errorf("unknown expression type %T", expr)
		}
	}
	return b.String(), nil
}

// Arg renders the bindvar at index.
func buildArg(ctx *context, index int) (string, error) {
	if index > len(ctx.Segment.Args) {
		return "", fmt.Errorf("invalid bindvar index %d", index)
	}
	i := index - 1
	ctx.ArgsUsed[i] = true
	built := ctx.ArgsBuilt[i]
	if built == "" || ctx.global.BindVarStyle == syntax.Question {
		*ctx.global.ArgStore = append(*ctx.global.ArgStore, ctx.Segment.Args[i])
		if ctx.global.BindVarStyle == syntax.Question {
			built = "?"
		} else {
			built = "$" + strconv.Itoa(len(*ctx.global.ArgStore))
		}
		ctx.ArgsBuilt[i] = built
	}
	return built, nil
}

// Column renders the column at index.
func buildColumn(ctx *context, index int) (string, error) {
	if index > len(ctx.Segment.Columns) {
		return "", fmt.Errorf("invalid column index %d", index)
	}
	i := index - 1
	ctx.ColumnsUsed[i] = true
	col := ctx.Segment.Columns[i]
	built := ctx.ColumnsBuilt[i]
	if built == "" || (ctx.global.BindVarStyle == syntax.Question && len(col.Args) > 0) {
		b, err := buildColumn2(ctx, col)
		if err != nil {
			return "", err
		}
		ctx.ColumnsBuilt[i] = b
		built = b
	}
	return built, nil
}

func buildColumn2(ctx *context, c *TableColumn) (string, error) {
	if c == nil || c.Raw == "" {
		return "", nil
	}
	seg := &Segment{
		Raw:    c.Raw,
		Args:   c.Args,
		Tables: []Table{c.Table},
	}
	ctx = newSegmentContext(ctx.global, seg)
	built, err := build(ctx)
	if err != nil {
		return "", err
	}
	// don't check usage of tables
	for i := range ctx.TableUsed {
		ctx.TableUsed[i] = true
	}
	if err := ctx.checkUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", ctx.Segment.Raw, err)
	}
	return built, err
}

func buildTable(ctx *context, index int) (string, error) {
	if index > len(ctx.Segment.Tables) {
		return "", fmt.Errorf("invalid table index %d", index)
	}
	ctx.TableUsed[index-1] = true
	return string(ctx.Segment.Tables[index-1]), nil
}

func buildSegment(ctx *context, index int) (string, error) {
	if index > len(ctx.Segment.Segments) {
		return "", fmt.Errorf("invalid segment index %d", index)
	}
	i := index - 1
	ctx.SegmentsUsed[i] = true
	seg := ctx.Segment.Segments[i]
	built := ctx.SegmentsBuilt[i]
	if built == "" || (ctx.global.BindVarStyle == syntax.Question && len(seg.Args) > 0) {
		b, err := seg.BuildContext(ctx.global)
		if err != nil {
			return "", err
		}
		ctx.SegmentsBuilt[i] = b
		built = b
	}
	return built, nil
}

func buildBuilder(ctx *context, index int) (string, error) {
	if index > len(ctx.Segment.Builders) {
		return "", fmt.Errorf("invalid builder index %d", index)
	}
	i := index - 1
	ctx.BuilderUsed[i] = true
	builder := ctx.Segment.Builders[i]
	built := ctx.BuildersBuilt[i]
	if built == "" || ctx.global.BindVarStyle == syntax.Question {
		b, err := builder.BuildContext(ctx.global)
		if err != nil {
			return "", err
		}
		ctx.BuildersBuilt[i] = b
		built = b
	}
	return built, nil
}
