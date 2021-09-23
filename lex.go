package JSONPath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrUnclosedString = errors.New("unclosed string literal")
	ErrBadEscape      = errors.New("unknown character escape")
	ErrShortEscape    = errors.New("unicode escape needs 4 hex digits")
)

// lexPath returns the next token and an optional associated value (eg, int or string), or an error.
// It interprets the tokens of path elements.
func lexPath(r *rd) (token, interface{}, error) {
	for isSpace(r.look()) {
		r.get()
	}
	switch c := r.get(); c {
	case eof:
		return tokEOF, nil, nil
	case '(', ')', '[', ']', '*', '$', ':', '-', ',': // - is allowed as a sign for integers
		return token(c), nil, nil
	case '.':
		return isNext(r, '.', tokNest, '.')
	case '?':
		return isNext(r, '(', tokFilter, tokError)
	case '"', '\'':
		s, err := lexString(r, c)
		return tokString, s, err
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
func lexExpr(r *rd, okRE bool) (token, interface{}, error) {
	for isSpace(r.look()) {
		r.get()
	}
	switch c := r.get(); c {
	case eof:
		return tokEOF, nil, nil
	case '(', ')', '[', ']', '@', '$', '.', ',':
		return token(c), nil, nil
	case '~', '*', '%', '+', '-':
		return token(c), nil, nil
	case '/':
		if okRE {
			s, err := lexString(r, c)
			return tokRE, s, err
		}
		return token(c), nil, nil
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
		return tokString, s, err
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
func lexNumber(r *rd) (token, int64, error) {
	var sb strings.Builder
	r.unget()
	for isDigit(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	v, err := strconv.ParseInt(sb.String(), 10, 64)
	if err != nil {
		return tokError, 0, err
	}
	return tokInt, v, nil
}

// lexID returns an identifier token from r
func lexID(r *rd) (token, string, error) {
	var sb strings.Builder
	r.unget()
	for isAlphanumeric(r.look()) {
		sb.WriteByte(byte(r.get()))
	}
	return tokID, sb.String(), nil
}

// lexString consumes a string from r until the closing quote cq, interpreting escape sequences.
func lexString(r *rd, cq int) (string, error) {
	var s strings.Builder
	for {
		c := r.get()
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
func isNext(r *rd, c int, t token, f token) (token, interface{}, error) {
	if r.look() == c {
		r.get()
		return t, nil, nil
	}
	if f == tokError {
		return tokError, r.look(), fmt.Errorf("unexpected char %q after %q at %s", rune(r.look()), rune(c), r.offset())
	}
	return f, nil, nil
}

// diagnose an unexpected character, not valid for a token
func tokenError(r *rd, c int) (token, int, error) {
	return tokError, c, fmt.Errorf("unexpected character %q at %s", rune(c), r.offset())
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
