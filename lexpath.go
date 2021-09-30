package JSONPath

import "math"

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
		return lexeme{tokString, s, err}
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
		n := fol.val.(int64)
		if n == math.MaxInt64 {
			return l.lexErr(ErrIntOverflow)
		}
		fol.val = -n
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
