package JSONPath

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrUnclosedString = errors.New("unclosed string literal")
	ErrBadEscape      = errors.New("unknown character escape sequence")
	ErrShortEscape    = errors.New("unicode escape needs 4 hex digits")
	ErrIntOverflow    = errors.New("overflow of negative integer literal")
	ErrBadReal        = errors.New("invalid floating-point literal syntax")
)

// lexeme is a tuple representing a lexical element: token, optional value, optional error
type lexeme struct {
	tok token
	val Val
	err error
	//	loc Loc	// starting location
}

// String returns a vaguely user-readable representation of a token
func (lx lexeme) String() string {
	if lx.err != nil {
		return "!" + lx.err.Error()
	}
	if lx.tok.hasVal() {
		return fmt.Sprintf("%v(%v)", lx.tok, lx.val)
	}
	return lx.tok.String()
}

// GoString returns a string with an internal representation of a token, for debugging.
func (lx lexeme) GoString() string {
	if lx.err != nil {
		return "!" + lx.err.Error()
	}
	if lx.tok.hasVal() {
		return fmt.Sprintf("%#v(%#v)", lx.tok, lx.val)
	}
	return lx.tok.GoString()
}

// lexer provides state and one-token lookahead for the token stream
type lexer struct {
	r    *rd
	peek bool   // unget was called
	lex  lexeme // value of unget
}

func newLexer(r *rd) *lexer {
	return &lexer{r: r}
}

// unget saves a lexeme.
func (l *lexer) unget(lex lexeme) {
	if l.peek {
		panic("internal error: too much lookahead")
	}
	l.peek = true
	l.lex = lex
}

// look saves a lexeme and returns its token part.
func (l *lexer) look(lx lexeme) token {
	l.unget(lx)
	if lx.err != nil {
		return tokError
	}
	return lx.tok
}

// offset returns a representation of the current stream offset
func (l *lexer) offset() string {
	return l.r.offset()
}

// lexErr returns a lexeme that bundles a diagnostic.
func (l *lexer) lexErr(err error) lexeme {
	return lexeme{tokError, nil, err}
}

// lexPath returns the next token and an optional associated value (eg, int or string), or an error.
// It interprets the tokens of path elements.
func (l *lexer) lexPath() lexeme {
	if l.peek {
		l.peek = false
		return l.lex
	}
	l.ws()
	r := l.r
	switch c := r.get(); c {
	case eof:
		return lexeme{tokEOF, nil, nil}
	case '(', ')', '[', ']', '*', '$', ':', ',':
		return lexeme{token(c), nil, nil}
	case '.':
		return l.isNext('.', tokNest, '.')
	case '?':
		return l.isNext('(', tokFilter, tokError)
	case '"', '\'':
		s, err := l.lexString(c)
		if err != nil {
			return l.lexErr(err)
		}
		return lexeme{tokString, StringVal(s), err}
	case '-': // - is allowed as a sign for integers
		l.ws()
		if !isDigit(r.get()) {
			r.unget()
			return l.tokenErr(r.look())
		}
		fol := l.lexNumber(false)
		if fol.tok != tokInt {
			return l.tokenErr(c)
		}
		n := fol.val.(IntVal)
		if n == math.MaxInt64 {
			return l.lexErr(ErrIntOverflow)
		}
		fol.val = IntVal(-n)
		return fol
	default:
		if isDigit(c) {
			return l.lexNumber(false)
		}
		if isLetter(c) {
			return l.lexID(isAlphanumericDash)
		}
		return l.tokenErr(c)
	}
}

// lexExpr returns the next token and an optional associated value (eg, int or string), or an error.
// It interprets the tokens of "script expressions" (filter and plain expressions).
func (l *lexer) lexExpr() lexeme {
	if l.peek {
		l.peek = false
		return l.lex
	}
	l.ws()
	r := l.r
	switch c := r.get(); c {
	case eof:
		return lexeme{tokEOF, nil, nil}
	case '(', ')', '[', ']', '@', '$', '.', ',', '~', '*', '%', '+', '-':
		return lexeme{token(c), nil, nil}
	case '/':
		return lexeme{token(c), nil, nil}
	case '&':
		return l.isNext('&', tokAnd, '&')
	case '|':
		return l.isNext('|', tokOr, '|')
	case '<':
		return l.isNext('=', tokLE, '<')
	case '>':
		return l.isNext('=', tokGE, '>')
	case '=':
		lx := l.isNext('=', tokEq, '=')
		if lx.tok != '=' {
			return lx
		}
		return l.isNext('~', tokMatch, '=')
	case '!':
		return l.isNext('=', tokNE, '!')
	case '"', '\'':
		s, err := l.lexString(c)
		if err != nil {
			return l.lexErr(err)
		}
		return lexeme{tokString, StringVal(s), err}
	default:
		if isDigit(c) {
			return l.lexNumber(true)
		}
		if isLetter(c) {
			return l.lexID(isAlphanumeric)
		}
		return l.tokenErr(c)
	}
}

// lexRegExp can be called by the parser when it consumes a token (eg, '/') that must introduce a regular expression,
// gathering the text of the expression here and returning it.
func (l *lexer) lexRegExp(c int) lexeme {
	s, err := l.lexString(c)
	return lexeme{tokRE, StringVal(s), err}
}

// ws skips white space, and returns the current location
func (l *lexer) ws() Loc {
	for isSpace(l.r.look()) {
		l.r.get()
	}
	return l.r.loc()
}

// lexNumber returns an integer token from r with a 64-bit value, or an error (eg, it overflows).
// Currently it supports only integers.
// The IETF grammar excludes leading zeroes, presumably to avoid octal, but we'll accept them as decimal.
func (l *lexer) lexNumber(real bool) lexeme {
	var sb strings.Builder
	r := l.r
	r.unget()
	for isDigit(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	if !real || r.look() != '.' {
		// integer only
		v, err := strconv.ParseInt(sb.String(), 10, 64)
		if err != nil {
			return l.lexErr(err)
		}
		return lexeme{tokInt, IntVal(v), nil}
	}
	r.get()
	for isDigit(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	if r.look() == 'e' { // e[+-]?[0-9]+
		r.get()
		c := r.look()
		if c == '+' || c == '-' {
			sb.WriteByte(byte(r.get()))
		}
		if !isDigit(r.look()) {
			return l.lexErr(ErrBadReal)
		}
		for isDigit(r.look()) {
			sb.WriteByte(byte(r.get()))
		}
	}
	v, err := strconv.ParseFloat(sb.String(), 64)
	if err != nil {
		return l.lexErr(err)
	}
	return lexeme{tokReal, FloatVal(v), nil}
}

// lexID returns an identifier token from r
func (l *lexer) lexID(isAlpha func(int) bool) lexeme {
	var sb strings.Builder
	r := l.r
	r.unget()
	for isAlpha(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	return lexeme{tokID, NameVal(sb.String()), nil}
}

// lexString consumes a string from r until the closing quote cq, interpreting escape sequences.
func (l *lexer) lexString(cq int) (string, error) {
	var s strings.Builder
	r := l.r
	for {
		c := r.get()
		if c == eof {
			return s.String(), ErrUnclosedString
		}
		if c == cq {
			break
		}
		if c == '\\' {
			r, err := escaped(r, cq)
			if err != nil {
				return s.String(), err
			}
			s.WriteRune(r)
		} else {
			s.WriteByte(byte(c))
		}
	}
	return s.String(), nil
}

// escaped returns the rune result of an escape sequence, or an error.
// cq is the closing quote character (only that one is allowed in a \ sequence).
// That seems a little fussy and might change.
func escaped(r *rd, cq int) (rune, error) {
	switch c := r.get(); c {
	case eof:
		return eof, ErrUnclosedString
	case '\\', '/':
		return rune(c), nil
	case 't':
		return '\t', nil
	case 'f':
		return 0xC, nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 'u':
		// unicode: exactly 4 hex digits (which isn't enough!)
		var digits strings.Builder
		var err error
	Hexed:
		for i := 0; i < 4; i++ {
			c = r.get()
			if c == eof {
				return eof, ErrUnclosedString
			}
			if !isHexDigit(c) {
				r.unget()
				// illegal unicode, what to do?
				err = ErrShortEscape
				// interpret the rune-so-far but also return an error
				break Hexed
			}
			digits.WriteByte(byte(c))
		}
		v, _ := strconv.ParseInt(digits.String(), 16, 32)
		return rune(v), err
	default:
		if c != cq {
			// illegal escape. what to do?
			r.unget()
			return '\\', ErrBadEscape
		}
		return rune(cq), nil
	}
}

// if the next character is c, consume it and return t, otherwise return f
func (l *lexer) isNext(c int, t token, f token) lexeme {
	r := l.r
	if r.look() == c {
		r.get()
		return lexeme{t, nil, nil}
	}
	if f == tokError {
		return l.lexErr(fmt.Errorf("unexpected char %q after %q at %s", rune(r.look()), rune(c), r.offset()))
	}
	return lexeme{f, nil, nil}
}

// diagnose an unexpected character, not valid for a token
func (l *lexer) tokenErr(c int) lexeme {
	return l.lexErr(fmt.Errorf("unexpected character %q at %s", rune(c), l.r.offset()))
}

func isDigit(c int) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c int) bool {
	return isDigit(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// All Unicode values outside the ASCII space are acceptable "letters", not just those acceptable to unicode.IsLetter.
// Note that c is actually a single byte from a UTF-8 stream.
func isLetter(c int) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c >= 0x80
}

func isAlphanumeric(c int) bool {
	return isLetter(c) || isDigit(c)
}

// weirdly, outside expressions, identifiers can contain "-".
func isAlphanumericDash(c int) bool {
	return isLetter(c) || isDigit(c) || c == '-'
}

// isSpace returns the space characters allowed by the grammar(s), and \f and \v.
// Note that the grammar assigns special Unicode spaces to the set of characters in identifiers.
func isSpace(c int) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
