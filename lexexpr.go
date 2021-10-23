package JSONPath

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
		lx := l.isNext('=', tokEQ, '=')
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
		return lexeme{tokString, s, err}
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

// lexRegexp can be called by the parser when it consumes a token (eg, '/') that must introduce a regular expression,
// gathering the text of the expression here and returning it.
func (l *lexer) lexRegexp(c int) lexeme {
	s, err := l.lexString(c)
	return lexeme{tokRE, s, err}
}
