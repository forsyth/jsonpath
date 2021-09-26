package JSONPath

// parser represents the state of the expression parser
type parser struct {
	*lexer // source of tokens
}
