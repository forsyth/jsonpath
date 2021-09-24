package JSONPath

import (
	"errors"
	"fmt"
)

type Val interface{}

type Path []*Step

type Step struct {
	op   Op
	args []Val
}

// started with the IETF drafts, but reverted to a grammar adapted from https://github.com/dchester/jsonpath/blob/master/lib/grammar.js
// path ::= "$" step*
// step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
// member ::= "*" | identifier | expr | integer
// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= integer
// array-slice ::= start? ":" end? (":" step?)?
// start ::= integer | expr
// end ::= integer | expr
// step ::= integer | expr
// expr ::= "(" script-expression ")"
// filter ::= "?(" script-expression ")"
// integer := "-"? [0-9]+

// ParsePath returns the parsed form of the path expression in s, or an error.
func ParsePath(s string) (Path, error) {
	parser := &parser{newLexer(&rd{s: s})}
	return parser.parsePath()
}

func (p *parser) lexPath() lexeme {
	return p.lexer.lexPath()
}

func (p *parser) lookPath() token {
	return p.lexer.look(p.lexer.lexPath())
}

func (p *parser) parsePath() (Path, error) {
	err := expect(p, '$')
	if err != nil {
		return nil, err
	}
	path := []*Step{&Step{OpRoot, nil}}
	for {
		lx := p.lexPath()
		switch lx.tok {
		case tokEOF:
			return path, nil
		case tokError:
			return nil, lx.err
		case '.':
			op, name, err := p.parseMember()
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				path = append(path, &Step{OpWild, nil})
			} else {
				path = append(path, &Step{op, []Val{name}})
			}
		case tokNest:
			if p.lookPath() == '[' {
				// ".." "[" subscript "]"
				p.lexPath()
				sub, err := p.parseBrackets()
				if err != nil {
					return nil, err
				}
				path = append(path, &Step{OpNest, []Val{sub}})
				break
			}
			op, name, err := p.parseMember()
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				return nil, errors.New("..* not allowed") // or is it?
			}
			path = append(path, &Step{OpNest, []Val{name}})
		case '[':
			sub, err := p.parseBrackets()
			if err != nil {
				return nil, err
			}
			path = append(path, sub)
		default:
			// TO DO: make lexeme printable
			if lx.tok.hasVal() {
				return nil, fmt.Errorf("unexpected token %v (%v)", lx.tok, lx.val)
			}
			return nil, fmt.Errorf("unexpected token %v", lx.tok)
		}
	}
}

// parse bracketed subscript in
// step ::= ...  "[" subscript "]" ... | ".." "[" subscript "]"
func (p *parser) parseBrackets() (*Step, error) {
	op, vals, err := p.parseSubscript()
	if err != nil {
		return nil, err
	}
	err = expect(p, ']')
	if err != nil {
		return nil, err
	}
	// op= OpSlice, OpIndex, OpSelect, OpUnion
	// vals structure distinguishes them
	return &Step{op, []Val{vals}}, nil
}

// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= integer
// array-slice ::= start? ":" end? (":" step?)?
func (p *parser) parseSubscript() (Op, Val, error) {
	union := []*Step{}	// if it's the sequence of union-element
	for {
		lx := p.lexPath()
		switch lx.tok {

		case tokError:
			return OpError, "", lx.err
		case tokEOF:
			return OpError, "", p.unexpectedEOF()

		// subscript-expression
		case '*':
			return OpWild, nil, nil
		case '(':
			e, err := p.parseExpr()
			// need to lookahead for ":", the "expr" case of "start" (ie, it's a slice)
			return OpExp, e, err
		case tokFilter:
			e, err := p.parseExpr()
			return OpFilter, e, err

		// union-element ("," union-element)
		// case '-': // TO DO: signed integer; easier to add it to lexPath
		case tokInt:
			// integer or start element of slice
			// need to lookahead for ":" (OpIndex vs OpSlice)
			union = append(union, &Step{OpIndex, []Val{lx.val}})
		case tokString:
			// string-literal
			union = append(union, &Step{OpString, []Val{lx.val}})
		case tokID:
			// treat same as string-literal
			union = append(union, &Step{OpId, []Val{lx.val}})
		// error
		default:
			return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, p.lexer.offset())
		}
		lx = p.lexPath()
		if lx.tok != ',' {
			if lx.tok != ']' {
				// TO DO: unclosed bracket
			}
			return OpUnion, union, nil
		}
	}
}

// member ::= "*" | identifier | expr | integer
func (p *parser) parseMember() (Op, Val, error) {
	lx := p.lexPath()
	if lx.err != nil {
		return OpError, "", lx.err
	}
	switch lx.tok {
	case '*':
		return OpWild, lx.val, nil
	case tokID:
		return OpId, lx.val, nil
	case tokString:
		return OpString, lx.val, nil
	case tokInt:
		return OpInt, lx.val, nil
	case '(':
		// expr ::= "(" script-expression ")"
		e, err := p.parseExpr()
		return OpExp, e, err
	default:
		return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, p.lexer.offset())
	}
}

// parse the tail of expr or filter, expecting a closing ')'
func (p *parser) parseExpr() (Expr, error) {
	e, err := p.parseScriptExpr()
	if err != nil {
		return nil, err
	}
	err = expect(p, ')')
	if err != nil {
		return nil, err
	}
	return e, nil
}

func expect(p *parser, nt token) error {
	lx := p.lexPath()
	if lx.err != nil {
		return lx.err
	}
	if lx.tok != nt {
		return fmt.Errorf("expected %q at %s, got %v", nt, p.lexer.offset(), lx.tok)
	}
	return nil
}

func (p *parser) unexpectedEOF() error {
	return fmt.Errorf("unexpected EOF at %s", p.lexer.offset())
}
