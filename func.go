package sqls

import (
	"fmt"
	"strconv"
	"strings"

	"git.qjebbs.com/jebbs/go-sqls/syntax"
)

func join(ctx *context, args ...string) (string, error) {
	if len(args) != 2 {
		return "", argError("column(tmpl, sep string)", args)
	}
	tmpl, separator := args[0], args[1]
	c, err := syntax.Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse enum template '%s': %w", tmpl, err)
	}
	b := new(strings.Builder)
	var (
		firstRefType string
		nRefs        int
		calls        []*syntax.FuncCallExpr
	)
	for i, expr := range c.ExprList {
		fn, ok := expr.(*syntax.FuncExpr)
		if !ok {
			continue
		}
		call := &syntax.FuncCallExpr{
			Name: fn.Name,
		}
		c.ExprList[i] = call
		calls = append(calls, call)
		switch call.Name {
		case "$", "?":
			if firstRefType == "" {
				firstRefType = "arg(s)"
				nRefs = len(ctx.s.Args)
			} else if nRefs != len(ctx.s.Args) {
				return "", fmt.Errorf("unaligned references %d '%s' to %d arg(s)", nRefs, firstRefType, len(ctx.s.Args))
			}
		case "c", "col", "column",
			"t", "table":
			if firstRefType == "" {
				firstRefType = "columns(s)"
				nRefs = len(ctx.s.Columns)
			} else if nRefs != len(ctx.s.Columns) {
				return "", fmt.Errorf("unaligned references %d '%s' to %d columns(s)", nRefs, firstRefType, len(ctx.s.Columns))
			}
		case "s", "seg", "segment":
			if firstRefType == "" {
				firstRefType = "segment(s)"
				nRefs = len(ctx.s.Segments)
			} else if nRefs != len(ctx.s.Segments) {
				return "", fmt.Errorf("unaligned references %d '%s' to %d segment(s)", nRefs, firstRefType, len(ctx.s.Segments))
			}
		}
	}
	if firstRefType == "" {
		return "", fmt.Errorf("no references found in enum template '%s'", tmpl)
	}
	for i := 0; i < nRefs; i++ {
		if i > 0 {
			b.WriteString(separator)
		}
		for _, call := range calls {
			call.Args = []string{strconv.Itoa(i + 1)}
		}
		s, err := build(ctx, c)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

func argumentDollar(ctx *context, args ...string) (string, error) {
	return arg(ctx, syntax.ArgIndexed, args...)
}

func argumentQuestion(ctx *context, args ...string) (string, error) {
	return arg(ctx, syntax.ArgUnindexed, args...)
}

func arg(ctx *context, typ syntax.RefType, args ...string) (string, error) {
	if ctx.bindVarStyle == "" {
		ctx.bindVarStyle = typ
		ctx.firstBindvar = ctx.s.Raw
	}
	if ctx.bindVarStyle != typ {
		return "", fmt.Errorf("mixed bindvar styles between segments '%s' and '%s'", ctx.firstBindvar, ctx.s.Raw)
	}
	if len(args) != 1 {
		switch ctx.bindVarStyle {
		case syntax.ArgIndexed:
			return "", argError("$(i int)", args)
		default:
			return "", argError("?(i int)", args)
		}
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid arg index '%s': %w", args[0], err)
	}
	return ctx.Arg(ctx, i)
}

func column(ctx *context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("column(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return ctx.Column(ctx, i)
}

func table(ctx *context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("table(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return ctx.Table(ctx, i)
}

func segment(ctx *context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("segment(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return ctx.Segment(ctx, i)
}

func argError(sig string, args any) error {
	return fmt.Errorf("bad args for #%s: got %v", sig, args)
}
