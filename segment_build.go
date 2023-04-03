package sqls

import (
	"fmt"
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
		newContext(argStore, s),
	)
}

func (s *Segment) buildInternal(ctx *context) (string, error) {
	if s == nil {
		return "", nil
	}
	body, err := s.buildBody(ctx)
	if err != nil {
		return "", err
	}
	if err := ctx.checkUsage(); err != nil {
		return "", fmt.Errorf("build '%s': %w", s.Raw, err)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return "", nil
	}
	header, footer := "", ""
	if s.Header != "" {
		header = s.Header + " "
	}
	if s.Footer != "" {
		footer = " " + s.Footer
	}
	return header + body + footer, nil
}

func (s *Segment) buildBody(ctx *context) (string, error) {
	if s == nil {
		return "", nil
	}
	clause, err := syntax.Parse(s.Raw)
	if err != nil {
		return "", fmt.Errorf("parse '%s': %w", s.Raw, err)
	}
	c, err := build(ctx, clause)
	if err != nil {
		return "", err
	}
	return c, err
}

func build(ctx *context, clause *syntax.Clause) (string, error) {
	b := new(strings.Builder)
	for _, decl := range clause.ExprList {
		switch expr := decl.(type) {
		case *syntax.PlainExpr:
			b.WriteString(expr.Text)
		case *syntax.FuncCallExpr:
			fn := ctx.funcMap[expr.Name]
			if fn == nil {
				return "", fmt.Errorf("function '%s' is not found", expr.Name)
			}
			s, err := fn(ctx, expr.Args...)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		case *syntax.RefExpr:
			switch expr.Type {
			case syntax.ArgIndexed, syntax.ArgUnindexed:
				if ctx.bindVarStyle == "" {
					ctx.bindVarStyle = expr.Type
					ctx.firstBindvar = ctx.s.Raw
				}
				if ctx.bindVarStyle != expr.Type {
					return "", fmt.Errorf("mixed bindvar styles between segments '%s' and '%s'", ctx.firstBindvar, ctx.s.Raw)
				}
				s, err := ctx.Arg(ctx, expr.Index)
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
