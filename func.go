package sqls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qjebbs/go-sqls/syntax"
)

// FuncMap is the map of preprocessing functions.
// Note: function names contain number [0-9] are not allowed.
type FuncMap map[string]Func

// Func is the type of preprocessing functions.
// A preprocessing function takes a ctx and args, returns the built string and any error.
//
// Two things for preprocessing functions to take care of, which are important:
//   - cache the built string and reuse it if it's possible, because rebuilding
//     leads to excessive argument allocation, and is expensive.
//   - set the used flags to indicate if an arg/column/table/segment is used,
//     since we need to check and make sure every arg/column/table/segment is used.
type Func func(ctx *Context, args ...string) (string, error)

func builtInFuncs() map[string]Func {
	return map[string]Func{
		"join":    join,
		"$":       argumentDollar,
		"?":       argumentQuestion,
		"c":       column,
		"col":     column,
		"column":  column,
		"t":       table,
		"table":   table,
		"s":       segment,
		"seg":     segment,
		"segment": segment,
	}
}

func join(ctx *Context, args ...string) (string, error) {
	if len(args) != 2 {
		return "", argError("join(tmpl, sep string)", args)
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
	nArgs := len(ctx.Segment.Args)
	nColumns := len(ctx.Segment.Columns)
	nSegments := len(ctx.Segment.Segments)
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
				nRefs = nArgs
			} else if nRefs != nArgs {
				return "", fmt.Errorf("unaligned references %d '%s' to %d arg(s)", nRefs, firstRefType, nArgs)
			}
		case "c", "col", "column",
			"tn", "tableName",
			"t", "ta", "tableAlias",
			"tna", "tableNameAlias":
			if firstRefType == "" {
				firstRefType = "columns(s)"
				nRefs = nColumns
			} else if nRefs != nColumns {
				return "", fmt.Errorf("unaligned references %d '%s' to %d columns(s)", nRefs, firstRefType, nColumns)
			}
		case "s", "seg", "segment":
			if firstRefType == "" {
				firstRefType = "segment(s)"
				nRefs = nSegments
			} else if nRefs != nSegments {
				return "", fmt.Errorf("unaligned references %d '%s' to %d segment(s)", nRefs, firstRefType, nSegments)
			}
		}
	}
	if firstRefType == "" {
		return "", fmt.Errorf("no references found in enum template '%s'", tmpl)
	}
	for i := 0; i < nRefs; i++ {
		for _, call := range calls {
			call.Args = []string{strconv.Itoa(i + 1)}
		}
		s, err := buildCluase(ctx, c)
		if err != nil {
			return "", err
		}
		if s != "" {
			if i > 0 {
				b.WriteString(separator)
			}
			b.WriteString(s)
		}
	}
	return b.String(), nil
}

func argumentDollar(ctx *Context, args ...string) (string, error) {
	return arg(ctx, syntax.Dollar, args...)
}

func argumentQuestion(ctx *Context, args ...string) (string, error) {
	return arg(ctx, syntax.Question, args...)
}

func arg(ctx *Context, typ syntax.BindVarStyle, args ...string) (string, error) {
	if ctx.Global.BindVarStyle == 0 {
		ctx.Global.BindVarStyle = typ
		// ctx.global.FirstBindvar = ctx.Segment.Raw
	}
	// if ctx.global.BindVarStyle != typ {
	// 	return "", fmt.Errorf("mixed bindvar styles between segments '%s' and '%s'", ctx.global.FirstBindvar, ctx.Segment.Raw)
	// }
	if len(args) != 1 {
		switch ctx.Global.BindVarStyle {
		case syntax.Dollar:
			return "", argError("$(i int)", args)
		default:
			return "", argError("?(i int)", args)
		}
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid arg index '%s': %w", args[0], err)
	}
	return buildArg(ctx, i)
}

func column(ctx *Context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("column(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return buildColumn(ctx, i)
}

func table(ctx *Context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("tableName(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return buildTable(ctx, i)
}

func segment(ctx *Context, args ...string) (string, error) {
	if len(args) != 1 {
		return "", argError("segment(i int)", args)
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid index '%s': %w", args[0], err)
	}
	return buildSegment(ctx, i)
}

func argError(sig string, args any) error {
	return fmt.Errorf("bad args for #%s: got %v", sig, args)
}
