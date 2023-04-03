package syntax

import (
	"unicode"
	"unicode/utf8"
)

// EOF is the EOF rune
const EOF rune = 0

// lexerHelper is the lexical analyzer
type lexerHelper struct {
	input     string // input string
	start     int    // byte position where current token start
	startPos  Pos    // position where current token start
	pos       int    // byte position of the current rune
	line, col uint   // line and column of the current rune
	rune      rune   // current rune
	width     int    // bytes width of the current rune
	prevWidth int    // bytes width of the previous rune
}

// newLexerHelper returns a new LexerHelper
func newLexerHelper(input string) *lexerHelper {
	l := &lexerHelper{
		input: input,
		line:  1,
	}
	l.Next() // initialize the first rune
	return l
}

// StartToken was called when emit a new token, who set the l.Start
// for the next token.
func (l *lexerHelper) StartToken(tokens ...any) {
	l.start = l.pos
	l.startPos = Pos{uint(l.line), uint(l.col)}
}

// Next moves to the next rune
func (l *lexerHelper) Next() rune {
	if l.pos >= len(l.input) {
		l.prevWidth = l.width
		l.width = 0
		l.rune = EOF
		return l.rune
	}
	l.pos += l.width
	result, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.prevWidth = l.width
	l.width = width
	l.rune = result
	if l.rune == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	return l.rune
}

// Peek returns the next rune without changing the postions.
// it returns the whole string left if n is 0.
func (l *lexerHelper) Peek() rune {
	next := l.pos + l.width
	if next >= len(l.input) {
		return EOF
	}
	r, _ := utf8.DecodeRuneInString(l.input[next:])
	return r
}

// PeekN returns the next n bytes string, without changing the postions.
// it returns the whole string left if n is 0.
func (l *lexerHelper) PeekN(n int) string {
	next := l.pos + l.width
	if next >= len(l.input) {
		return ""
	}
	if n == 0 {
		return l.input[next:]
	}
	return l.input[next : next+n]
}

// Back moves back to the previous rune
func (l *lexerHelper) Back() {
	l.pos -= l.prevWidth
}

// SkipWhitespace skips all leading whitespaces
func (l *lexerHelper) SkipWhitespace() {
	for {
		if !unicode.IsSpace(l.rune) {
			break
		}
		l.Next()
		if l.rune == EOF {
			break
		}
	}
}

// IsEOF tells if it reaches the EOF
func (l *lexerHelper) IsEOF() bool {
	return l.pos >= len(l.input)
}

// IsWhitespace tells if it's currently a whitespace
func (l *lexerHelper) IsWhitespace() bool {
	return unicode.IsSpace(l.rune)
}

// Advanced tells if current position is advanced compared to the start position
func (l *lexerHelper) Advanced() bool {
	return l.pos > l.start
}

// Lower returns lower-case ch iff ch is ASCII letter
func (l *lexerHelper) Lower() rune {
	return ('a' - 'A') | l.rune
}

func (l *lexerHelper) IsLetter() bool {
	return 'a' <= l.Lower() && l.Lower() <= 'z' || l.rune == '_'
}

func (l *lexerHelper) IsDecimal() bool {
	return '0' <= l.rune && l.rune <= '9'
}

func (l *lexerHelper) IsHex() bool {
	return '0' <= l.rune && l.rune <= '9' || 'a' <= l.Lower() && l.Lower() <= 'f'
}
