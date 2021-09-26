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
	ErrBadEscape      = errors.New("unknown character escape")
	ErrShortEscape    = errors.New("unicode escape needs 4 hex digits")
	ErrIntOverflow    = errors.New("overflow of negative integer literal")
)

// lexeme is a tuple representing a lexical element: token, optional value, optional error
type lexeme struct {
	tok token
	val Val
	err error
}

// String returns a vaguely user-readable representation of a token
func (lx lexeme) String() string {
	if lx.err != nil {
		return lx.err.Error()
	}
	if lx.tok.hasVal() {
		return fmt.Sprintf("%v(%v)", lx.tok, lx.val)
	}
	return lx.tok.String()
}

// GoString returns a string with an internal representation of a token, for debugging.
func (lx lexeme) GoString() string {
	if lx.err != nil {
		return lx.err.Error()
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

// lexPath returns the next token and an optional associated value (eg, int or string), or an error.
// It interprets the tokens of path elements.
func (l *lexer) lexPath() lexeme {
	if l.peek {
		l.peek = false
		return l.lex
	}
	r := l.r
	for isSpace(r.look()) {
		r.get()
	}
	switch c := r.get(); c {
	case eof:
		return lexeme{tokEOF, nil, nil}
	case '(', ')', '[', ']', '*', '$', ':', ',':
		return lexeme{token(c), nil, nil}
	case '.':
		return isNext(r, '.', tokNest, '.')
	case '?':
		return isNext(r, '(', tokFilter, tokError)
	case '"', '\'':
		s, err := lexString(r, c)
		return lexeme{tokString, s, err}
	case '-': // - is allowed as a sign for integers
		fol := l.lexPath()
		if fol.tok != tokInt {
			return tokenError(r, c)
		}
		n := fol.val.(int64)
		if n == math.MaxInt64 {
			return lexeme{tokError, "", ErrIntOverflow}
		}
		fol.val = -n
		return fol
	default:
		if isDigit(c) {
			return lexNumber(r)
		}
		if isLetter(c) {
			return lexID(r)
		}
		return tokenError(r, c)
	}
}

// lexExpr returns the next token and an optional associated value (eg, int or string), or an error.
// It interprets the tokens of "script expressions" (filter and plain expressions).
// If okRE is true, a '/' character introduces a regular expression (ended by an unescaped trailing '/').
func (l *lexer) lexExpr(okRE bool) lexeme {
	if l.peek {
		l.peek = false
		return l.lex
	}
	r := l.r
	for isSpace(r.look()) {
		r.get()
	}
	switch c := r.get(); c {
	case eof:
		return lexeme{tokEOF, nil, nil}
	case '(', ')', '[', ']', '@', '$', '.', ',':
		return lexeme{token(c), nil, nil}
	case '~', '*', '%', '+', '-':
		return lexeme{token(c), nil, nil}
	case '/':
		if okRE {
			s, err := lexString(r, c)
			return lexeme{tokRE, s, err}
		}
		return lexeme{token(c), nil, nil}
	case '&':
		return isNext(r, '&', tokAnd, '&')
	case '|':
		return isNext(r, '|', tokOr, '|')
	case '<':
		return isNext(r, '=', tokLE, '<')
	case '>':
		return isNext(r, '=', tokGE, '>')
	case '=':
		return isNext(r, '=', tokEq, '=')
	case '!':
		return isNext(r, '=', tokNE, '!')
	case '"', '\'':
		s, err := lexString(r, c)
		return lexeme{tokString, s, err}
	default:
		if isDigit(c) {
			return lexNumber(r)
		}
		if isLetter(c) {
			return lexID(r)
		}
		return tokenError(r, c)
	}
}

// lexNumber returns an integer token from r with a 64-bit value, or an error (eg, it overflows).
// Currently it supports only integers.
// The IETF grammar excludes leading zeroes, presumably to avoid octal, but we'll accept them as decimal.
func lexNumber(r *rd) lexeme {
	var sb strings.Builder
	r.unget()
	for isDigit(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	v, err := strconv.ParseInt(sb.String(), 10, 64)
	if err != nil {
		return lexeme{tokError, 0, err}
	}
	return lexeme{tokInt, v, nil}
}

// lexID returns an identifier token from r
func lexID(r *rd) lexeme {
	var sb strings.Builder
	r.unget()
	for isAlphanumeric(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	return lexeme{tokID, sb.String(), nil}
}

// lexString consumes a string from r until the closing quote cq, interpreting escape sequences.
func lexString(r *rd, cq int) (string, error) {
	var s strings.Builder
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
func isNext(r *rd, c int, t token, f token) lexeme {
	if r.look() == c {
		r.get()
		return lexeme{t, nil, nil}
	}
	if f == tokError {
		return lexeme{tokError, r.look(), fmt.Errorf("unexpected char %q after %q at %s", rune(r.look()), rune(c), r.offset())}
	}
	return lexeme{f, nil, nil}
}

// diagnose an unexpected character, not valid for a token
func tokenError(r *rd, c int) lexeme {
	return lexeme{tokError, c, fmt.Errorf("unexpected character %q at %s", rune(c), r.offset())}
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

// isSpace returns the space characters allowed by the grammar(s), and \f and \v.
// Note that the grammar assigns special Unicode spaces to the set of characters in identifiers.
func isSpace(c int) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
