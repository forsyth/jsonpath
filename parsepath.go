package JSONPath

import (
	"errors"
	"fmt"
)

// path ::= "$" step*
// step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
// member ::= "*" | identifier | expr | signed-integer
// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= signed-integer
// array-slice ::= start? ":" end? (":" stride?)?
// start ::= signed-integer | expr
// end ::= signed-integer | expr
// stride ::= signed-integer | expr
// expr ::= "(" script-expression ")"
// filter ::= "?(" script-expression ")"
// signed-integer := "-"? integer
// integer ::= [0-9]+

// ParsePath returns the parsed form of the path expression in s, or an error.
func ParsePath(s string) (Path, error) {
	return newParser(s).parsePath()
}

func (p *parser) lookPath() token {
	return p.look(p.lexPath())
}

// path ::= "$" step*
// step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
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
			return nil, fmt.Errorf("unexpected token %v", lx)
		}
	}
}

// parse bracketed subscript in
// step ::= ...  "[" subscript "]" ... | ".." "[" subscript "]"
func (p *parser) parseBrackets() (*Step, error) {
	step, err := p.parseSubscript()
	if err != nil {
		return nil, err
	}
	err = expect(p, ']')
	if err != nil {
		return nil, err
	}
	return step, nil
}

// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= signed-integer
// array-slice ::= start? ":" end? (":" stride?)?
//
// it's easier to accept a list of any both subscript-expressions and union-elements
// and analyse the value list to see what it is (or if it's an illegal list)
func (p *parser) parseSubscript() (*Step, error) {
	steps, err := p.parseValList()
	if err != nil {
		return nil, err
	}
	if len(steps) > 1 { // can't have subscript-expressions
		for _, step := range steps {
			switch step.Op {
			case OpWild, OpExp, OpFilter:
				return nil, fmt.Errorf("%s cannot be in a union element list", step.Op)
			default:
				// ok
			}
		}
	}
	// distinguish union from subscript expression
	switch steps[0].Op {
	case OpWild, OpFilter:
		return steps[0], nil
	case OpExp:
		steps[0].Op = OpIndex
		return steps[0], nil
	default:
		args := []Val{}
		for _, step := range steps {
			if len(step.Args) != 1 {
				panic(fmt.Sprintf("incorrect structure for subscript list: %s", step.Op))
			}
			args = append(args, step.Args[0])
		}
		if len(args) == 1 {
			return &Step{OpSelect, args}, nil
		}
		return &Step{OpUnion, args}, nil
	}
}

// element ("," element)*
// where element ::= union-element | subscript-expression,
// where the latter cannot appear in a list.
// parseValList uses steps as a return value to tag the values with their internal type.
func (p *parser) parseValList() ([]*Step, error) {
	vals := []*Step{}
	for {
		exp, err := p.parseVal()
		if err != nil {
			return nil, err
		}
		vals = append(vals, exp)
		if p.lookPath() != ',' {
			break
		}
		p.lexPath()
	}
	return vals, nil
}

// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= signed-integer
// array-slice ::= start? ":" end? (":" stride?)?
func (p *parser) parseVal() (*Step, error) {
	lx := p.lexPath()
	if lx.err != nil {
		return nil, lx.err
	}
	switch lx.tok {
	case tokError:
		return nil, lx.err
	case tokEOF:
		return nil, p.unexpectedEOF()

	// subscript-expression
	case '*':
		return &Step{OpWild, nil}, nil
	case '(':
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.lookPath() == ':' {
			p.lexPath()
			return p.parseSlice(e)
		}
		return &Step{OpExp, []Val{e}}, nil
	case ':':
		// slice with missing start value
		return p.parseSlice(nil)

	case tokFilter:
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &Step{OpFilter, []Val{e}}, nil

	// definitely union-element
	case tokInt:
		// integer or start element of slice
		// need to lookahead for ":" (OpIndex vs OpSlice)
		if p.lookPath() == ':' {
			p.lexPath()
			return p.parseSlice(lx.val.(int64))
		}
		return &Step{OpInt, []Val{lx.val}}, nil
	case tokString:
		// string-literal
		return &Step{OpString, []Val{lx.val}}, nil
	case tokID:
		// treat same as string-literal
		return &Step{OpId, []Val{lx.val}}, nil

	default:
		// illegal
		return nil, fmt.Errorf("unexpected %v at %s", lx.tok, p.offset())
	}
}

// array-slice ::= start? ":" end? (":" stride?)?
// The initial start? ":" has been consumed, the ":" alerting us to a slice.
// The context means that valid successors are end, ":" (if a stride), "]", "," (if a union with a slice, and no stride),
// and then "]" if there's no stride, or there's a stride expression.
func (p *parser) parseSlice(start Val) (*Step, error) {
	slice := &Slice{start, nil, nil}
	step := &Step{OpSlice, []Val{slice}}
	tok := p.lookPath()
	if tok == ',' || tok == ']' {
		// neither end nor stride, ie x[s:], make them empty
		return step, nil
	}
	// end? (":" stride?)?
	if tok != ':' {
		// must be end (":" stride?)?
		e, err := p.parseSliceVal()
		if err != nil {
			return nil, err
		}
		slice.End = e
	}
	// (":" stride?)?
	switch p.lookPath() {
	case ']', ',':
		return step, nil
	case ':':
		// ":" stride?
		p.lexPath()
		tok = p.lookPath()
		if tok == ',' || tok == ']' {
			return step, nil
		}
		e, err := p.parseSliceVal()
		if err != nil {
			return nil, err
		}
		slice.Stride = e
		return step, nil
	default:
		return nil, fmt.Errorf("unexpected token %s at %s", p.lookPath(), p.offset())
	}
}

// (end|stride) ::= signed-integer | expr
func (p *parser) parseSliceVal() (Val, error) {
	switch lx := p.lexPath(); lx.tok {
	case '(':
		return p.parseExpr()
	case tokInt:
		return lx.val.(int64), nil
	default:
		return nil, fmt.Errorf("unexpected %v at %s", lx.tok, p.offset())
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
		return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, p.offset())
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
		return fmt.Errorf("expected %q at %s, got %v", nt, p.offset(), lx.tok)
	}
	return nil
}

func (p *parser) unexpectedEOF() error {
	return fmt.Errorf("unexpected EOF at %s", p.offset())
}
