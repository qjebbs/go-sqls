package syntax

// Clause is the clause.
type Clause struct {
	ExprList []Expr
}

// Expr is the declaration.
type Expr interface {
	Node
	aExpr()
}

// Node is the node.
type Node interface {
	Pos() Pos
	aNode()
}

type expr struct {
	node
}

func (*expr) aExpr() {}

type node struct {
	pos Pos
}

func (n *node) Pos() Pos { return n.pos }
func (*node) aNode()     {}

// BindVarExpr is the reference declaration.
type BindVarExpr struct {
	Type  BindVarStyle
	Index int
	expr
}

// BindVarStyle is the type of placeholder.
type BindVarStyle int

const (
	// Auto detect bindvar style, first encountered style will be used.
	Auto BindVarStyle = iota
	// Dollar is the type of indexed argument placeholders, e.g.: $1, $2, $3
	Dollar
	// Question is the type of unindexed argument placeholders, e.g.: ?, ?, ?
	Question
)

// FuncCallExpr is the function calling declaration.
type FuncCallExpr struct {
	Name string
	Args []string
	expr
}

// FuncExpr is the function declaration.
type FuncExpr struct {
	Name string
	expr
}

// PlainExpr is the plain text declaration.
type PlainExpr struct {
	Text string
	expr
}
