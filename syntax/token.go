package syntax

type token struct {
	typ   TokenType
	lit   string
	bad   bool
	kind  litKind
	start int
	end   int
	pos   Pos
}

// TokenType is the type of token.
type TokenType string

const (
	_EOF TokenType = "EOF"

	_Ref     = "ref"
	_Name    = "name"
	_Literal = "literal"
	_Plain   = "plain text"

	// delimiter
	_Hash   = "#"
	_Lparen = "("
	_Rparen = ")"
	_Comma  = ","
)

type litKind uint8

const (
	_IntLit litKind = iota
	_FloatLit
	_StringLit
	_BoolLit
	_NilLit
)
