package syntax

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses the input and returns the list of expressions.
func Parse(input string) (*Clause, error) {
	p := &parser{
		scanner: newScanner(input),
	}
	if err := p.Parse(); err != nil {
		return nil, err
	}
	return p.c, nil
}

type parser struct {
	*scanner

	bindVarIndex int
	bindVarStyle BindVarStyle
	// buf []token

	c *Clause
}

func (p *parser) want(t TokenType) error {
	if !p.got(t) {
		return p.syntaxError(
			fmt.Sprintf("syntax error: unexpected %s, want %s", p.token.typ, t),
		)
	}
	return nil
}

func (p *parser) got(tok TokenType) bool {
	p.NextToken()
	if p.token.typ == tok {
		return true
	}
	return false
}

func (p *parser) syntaxError(msg string) error {
	return fmt.Errorf("%d: syntax error: %s", p.token.start, msg)
}

func (p *parser) Parse() error {
	p.c = &Clause{}
	for p.NextToken() {
		switch p.token.typ {
		case _EOF:
			break
		case _Ref:
			d, err := p.refExpr()
			if err != nil {
				return err
			}
			p.c.ExprList = append(p.c.ExprList, d)
		case _Hash:
			err := p.funcExpr()
			if err != nil {
				return err
			}
		case _Plain:
			p.c.ExprList = append(p.c.ExprList, &PlainExpr{Text: p.token.lit})
		default:
			return p.syntaxError("unexpected token " + string(p.token.typ))
		}
	}
	return nil
}

func (p *parser) refExpr() (Expr, error) {
	pos := p.token.pos
	var t BindVarStyle
	switch p.token.lit {
	case "$":
		t = Dollar
		p.bindVarIndex++
		if p.bindVarStyle == 0 {
			p.bindVarStyle = t
		}
		if p.bindVarStyle != t {
			return nil, p.syntaxError("mixed bindvar styles")
		}
	case "?":
		t = Question
		p.bindVarIndex++
		if p.bindVarStyle == 0 {
			p.bindVarStyle = t
		}
		if p.bindVarStyle != t {
			return nil, p.syntaxError("mixed bindvar styles")
		}
	}
	index := p.bindVarIndex
	if t != Question {
		if err := p.want(_Literal); err != nil {
			return nil, err
		}
		if p.token.kind != _IntLit {
			return nil, p.syntaxError("unexpected '" + p.token.lit + "', want bindvar index")
		}
		val, err := strconv.ParseUint(p.token.lit, 10, 64)
		if err != nil {
			return nil, p.syntaxError(err.Error())
		}
		index = int(val)
	}
	return &BindVarExpr{
		Type:  t,
		Index: index,
		expr:  expr{node{pos}},
	}, nil
}

func (p *parser) funcExpr() error {
	pos := p.token.pos
	if err := p.want(_Name); err != nil {
		return err
	}
	nameToken := p.token
	p.NextToken()
	switch p.token.typ {
	case _Lparen:
		args := []string{}
		for {
			if !p.got(_Literal) {
				if p.token.typ != _Rparen {
					return p.syntaxError("unexpected token " + string(p.token.typ) + ", want args")
				}
				break
			}
			if p.token.bad {
				return p.syntaxError("bad argument: " + p.token.lit)
			}
			arg := p.token.lit
			if p.token.kind == _StringLit {
				arg = strings.ReplaceAll(arg[1:len(arg)-1], "''", "'")
			}
			args = append(args, arg)
			if !p.got(_Comma) {
				break
			}
		}
		if p.token.typ != _Rparen {
			return p.syntaxError("unexpected token " + string(p.token.typ) + ", want )")
		}
		p.c.ExprList = append(
			p.c.ExprList,
			&FuncCallExpr{
				Name: nameToken.lit,
				Args: args,
				expr: expr{node{pos}},
			},
		)
	case _Literal:
		p.c.ExprList = append(
			p.c.ExprList,
			&FuncCallExpr{
				Name: nameToken.lit,
				Args: []string{p.token.lit},
				expr: expr{node{pos}},
			},
		)
	default:
		p.c.ExprList = append(
			p.c.ExprList,
			&FuncExpr{Name: nameToken.lit},
			&PlainExpr{Text: p.token.lit},
		)
	}

	return nil
}
