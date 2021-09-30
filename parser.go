package JSONPath

import "fmt"

// parser represents the state of the path and/or expression parser
// (all of which is currently in the lexer), and provides a scope for the
// parsing methods.
type parser struct {
	*lexer // source of tokens
}

// newParser initialises and returns a parser
func newParser(s string) *parser {
	return &parser{newLexer(&rd{s: s})}
}

func (p *parser) expect(lex func () lexeme, nt token) error {
	lx := lex()
	if lx.err != nil {
		return lx.err
	}
	if lx.tok != nt {
		return fmt.Errorf("expected %q at %s, got %v", nt, p.offset(), lx.tok)
	}
	return nil
}
