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

// RefExpr is the reference declaration.
type RefExpr struct {
	Type  RefType
	Index int
	expr
}

// RefType is the type of placeholder.
type RefType string

const (
	// ArgIndexed is the type of indexed argument placeholders, e.g.: $1, $2, $3
	ArgIndexed RefType = "indexed bindvar"
	// ArgUnindexed is the type of unindexed argument placeholders, e.g.: ?, ?, ?
	ArgUnindexed RefType = "unindexed bindvar"
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
