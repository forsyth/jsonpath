package JSONPath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

var (
	ErrUnclosedString = errors.New("unclosed string literal")
	ErrBadEscape      = errors.New("unknown character escape sequence")
	ErrShortEscape    = errors.New("unicode escape needs 4 hex digits")
	ErrIntOverflow    = errors.New("overflow of negative integer literal")
	ErrBadReal        = errors.New("invalid floating-point literal syntax")
	ErrCtrlChar	= errors.New("must use escape to encode ctrl character")
)

// lexeme is a tuple representing a lexical element: token, optional value, optional error
type lexeme struct {
	tok token
	val interface{}
	err error
	//	loc Loc	// starting location
}

// String returns a vaguely user-readable representation of a token.
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

// s returns the value of the token as a string, or panics.
func (lx lexeme) s() string {
	return lx.val.(string)
}

// i returns the value of the token as an integer, or panics.
func (lx lexeme) i() int64 {
	return lx.val.(int64)
}

// f returns the value of the token as floating-point, or panics.
func (lx lexeme) f() float64 {
	return lx.val.(float64)
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
		return lexeme{tokInt, v, nil}
	}
	sb.WriteByte(byte(r.get()))
	for isDigit(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	if r.look() == 'e' || r.look() == 'E' { // e[+-]?[0-9]+
		sb.WriteByte(byte(r.get()))
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
	return lexeme{tokReal, v, nil}
}

// lexID returns an identifier token from r
func (l *lexer) lexID(isAlpha func(int) bool) lexeme {
	var sb strings.Builder
	r := l.r
	r.unget()
	for isAlpha(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	return lexeme{tokID, sb.String(), nil}
}

// lexString consumes a string from r until the closing quote cq, interpreting escape sequences,
// including the wretched surrogate pairs as escape sequences.
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
			// need to decode surrogate pairs to match JSON's rules
			surr := rune(0)
			for ; c == '\\'; c = r.get() {
				rv, err := escaped(r, cq)
				if err != nil {
					return s.String(), err
				}
				if surr != 0 {
					// previous \u was surrogate, if this was \u too match it up
					// note that if rv wasn't \u it won't be a valid successor.
					paired := utf16.DecodeRune(surr, rv)
					s.WriteRune(paired)
					surr = 0
					if paired != unicode.ReplacementChar {
						// consume surrogate and its pair
						continue
					}
					// invalid surrogate pair: already produced ReplacementChar
				} else if utf16.IsSurrogate(rv) {
					surr = rv
					continue
				}
				s.WriteRune(rv)
			}
			r.unget()
			if surr != 0 {
				// surrogate prefix without successor
				s.WriteRune(unicode.ReplacementChar)
			}
		} else {
			if c < 0x20 {
				// cannot include control characters directly
				return s.String(), ErrCtrlChar
			}
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
	case 'b':
		return '\b', nil
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
		v, _ := strconv.ParseUint(digits.String(), 16, 32)
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
