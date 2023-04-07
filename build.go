package sqls

import (
	"fmt"
	"strconv"
	"strings"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

// Build builds the segment.
func (s *Segment) Build() (query string, args []any, err error) {
	args = make([]any, 0)
	query, err = s.BuildTo(&args)
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

// BuildTo builds the segment to the arg storage.
func (s *Segment) BuildTo(argStore *[]any) (query string, err error) {
	return s.buildInternal(
		newContext(argStore),
	)
}

func (s *Segment) buildInternal(ctx *context) (string, error) {
	if s == nil {
		return "", nil
	}
	ctx = ctx.WithSegment(s)
	body, err := build(ctx)
	if err != nil {
		return "", err
	}
	if err := ctx.Current.checkUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", ctx.Current.Segment.Raw, err)
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
	clause, err := syntax.Parse(ctx.Current.Segment.Raw)
	if err != nil {
		return "", fmt.Errorf("parse '%s': %w", ctx.Current.Segment.Raw, err)
	}
	built, err := buildCluase(ctx, clause)
	if err != nil {
		return "", fmt.Errorf("build '%s': %w", ctx.Current.Segment.Raw, err)
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
			fn := ctx.Global.FuncMap[expr.Name]
			if fn == nil {
				return "", fmt.Errorf("function '%s' is not found", expr.Name)
			}
			s, err := fn(ctx, expr.Args...)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		case *syntax.BindVarExpr:
			if ctx.Global.BindVarStyle == 0 {
				ctx.Global.BindVarStyle = expr.Type
				// ctx.Global.FirstBindvar = ctx.Current.Segment.Raw
			}
			switch expr.Type {
			case syntax.BindVarDollar, syntax.BindVarQuestion:
				// if ctx.Global.BindVarStyle != expr.Type {
				// 	return "", fmt.Errorf("mixed bindvar styles between segments '%s' and '%s'", ctx.Global.FirstBindvar, ctx.Current.Segment.Raw)
				// }
				s, err := buildArg(ctx, expr.Index)
				if err != nil {
					return "", err
				}
				b.WriteString(s)
			}
		default:
			return "", fmt.Errorf("unknown expression type %T", expr)
		}
	}
	return b.String(), nil
}

// Arg renders the bindvar at index.
func buildArg(ctx *context, index int) (string, error) {
	if index > len(ctx.Current.Segment.Args) {
		return "", fmt.Errorf("invalid bindvar index %d", index)
	}
	i := index - 1
	ctx.Current.ArgsUsed[i] = true
	built := ctx.Current.ArgsBuilt[i]
	if built == "" || ctx.Global.BindVarStyle == syntax.BindVarQuestion {
		*ctx.Global.ArgStore = append(*ctx.Global.ArgStore, ctx.Current.Segment.Args[i])
		if ctx.Global.BindVarStyle == syntax.BindVarQuestion {
			built = "?"
		} else {
			built = "$" + strconv.Itoa(len(*ctx.Global.ArgStore))
		}
		ctx.Current.ArgsBuilt[i] = built
	}
	return built, nil
}

// Column renders the column at index.
func buildColumn(ctx *context, index int) (string, error) {
	if index > len(ctx.Current.Segment.Columns) {
		return "", fmt.Errorf("invalid column index %d", index)
	}
	i := index - 1
	ctx.Current.ColumnsUsed[i] = true
	col := ctx.Current.Segment.Columns[i]
	built := ctx.Current.ColumnsBuilt[i]
	if built == "" || (ctx.Global.BindVarStyle == syntax.BindVarQuestion && len(col.Args) > 0) {
		b, err := buildColumn2(ctx, col)
		if err != nil {
			return "", err
		}
		ctx.Current.ColumnsBuilt[i] = b
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
	ctx = ctx.WithSegment(seg)
	built, err := build(ctx)
	if err != nil {
		return "", err
	}
	// don't check usage of tables
	for i := range ctx.Current.TableUsed {
		ctx.Current.TableUsed[i] = true
	}
	if err := ctx.Current.checkUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", ctx.Current.Segment.Raw, err)
	}
	return built, err
}

// Table renders the table at index.
func buildTable(ctx *context, index int) (string, error) {
	if index > len(ctx.Current.Segment.Tables) {
		return "", fmt.Errorf("invalid table index %d", index)
	}
	ctx.Current.TableUsed[index-1] = true
	return string(ctx.Current.Segment.Tables[index-1]), nil
}

// Segment renders the segment at index.
func buildSegment(ctx *context, index int) (string, error) {
	if index > len(ctx.Current.Segment.Segments) {
		return "", fmt.Errorf("invalid segment index %d", index)
	}
	i := index - 1
	ctx.Current.SegmentsUsed[i] = true
	seg := ctx.Current.Segment.Segments[i]
	built := ctx.Current.SegmentsBuilt[i]
	if built == "" || (ctx.Global.BindVarStyle == syntax.BindVarQuestion && len(seg.Args) > 0) {
		b, err := seg.buildInternal(ctx)
		if err != nil {
			return "", err
		}
		ctx.Current.SegmentsBuilt[i] = b
		built = b
	}
	return built, nil
}
