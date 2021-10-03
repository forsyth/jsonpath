package JSONPath

// Parse the script expression language embedded within path expressions.
// It's a small subset of JavaScript (at least I hope it's a small subset.)

import (
	//	"errors"
	"fmt"
)

// ParseScriptExpression gives direct access to the secondary parser for expressions, returning an Expr tree representing
// the expression in s.
func ParseScriptExpression(s string) (Expr, error) {
	p := newParser(s)
	e, err := p.parseScriptExpr()
	if err != nil {
		return nil, err
	}
	lx := p.lexExpr()
	if lx.tok != tokEOF {
		return nil, fmt.Errorf("missing operator at %s, before %s", p.offset(), lx.tok)
	}
	return e, nil
}

// lookExpr looks ahead in the expression lexical syntax.
func (p *parser) lookExpr() token {
	return p.look(p.lexExpr())
}

// advanceExpr consumes a token in the expression lexical syntax.
func (p *parser) advanceExpr() {
	_ = p.lexExpr()
}

// parseScriptExpr consumes and parses a script expression, using the expression lexical syntax (lexExpr).
func (p *parser) parseScriptExpr() (Expr, error) {
	return p.expr(0)
}

// expr collects binary operators with priority >= pri, starting with an initial primary tree:
//	primary (op e)*
// See http://antlr.org/papers/Clarke-expr-parsing-1986.pdf for the history and details.
// p.expr(0) builds a complete (sub)tree.
func (p *parser) expr(pri int) (Expr, error) {
	e, err := p.primary()
	if err != nil {
		return nil, err
	}
	// build tree nodes until a lower-priority operator is seen (including all non-binary-operators)
	for tok2op(p.lookExpr()).precedence() >= pri {
		lx := p.lexExpr()
		if lx.err != nil {
			return nil, lx.err
		}
		op := tok2op(lx.tok)
		oprec := op.precedence()
		right, err := p.expr(oprec + op.associativity())
		if err != nil {
			return nil, err
		}
		e = &Inner{op, []Expr{e, right}}
	}
	return e, nil
}

// unary applies a unary operator to a following primary expression
// unary ::= ("-" | "!")+ primary
func (p *parser) unary(op Op) (Expr, error) {
	arg, err := p.primary()
	if err != nil {
		return nil, err
	}
	return &Inner{op, []Expr{arg}}, nil
}

// primary ::= primary1 ("(" e-list ")" | "[" e "]" | "." identifier)*
func (p *parser) primary() (Expr, error) {
	e, err := p.primary1()
	if err != nil {
		return nil, err
	}
	for {
		switch p.lookExpr() {
		case '(':
			// function call
			p.advanceExpr()
			e, err = p.application(OpCall, ')', e)
			if err != nil {
				return nil, err
			}
		case '[':
			// index, just one expression
			p.advanceExpr()
			index, err := p.expr(0)
			if err != nil {
				return nil, err
			}
			err = p.expect(p.lexExpr, ']')
			if err != nil {
				return nil, err
			}
			return &Inner{OpIndex, []Expr{e, index}}, nil
		case '.':
			// field selection
			p.advanceExpr()
			lx := p.lexExpr()
			if lx.err != nil {
				return nil, lx.err
			}
			if lx.tok != tokID {
				return nil, fmt.Errorf("expected identifier in '.' selection")
			}
			e = &Inner{OpSelect, []Expr{e, &NameLeaf{OpID, lx.s()}}}
		default:
			return e, nil
		}
	}
}

// apply optional expression e to an expression list (terminated by a given end token) as operator op
func (p *parser) application(op Op, end token, e Expr) (Expr, error) {
	args, err := p.parseExprList(e)
	if err != nil {
		return nil, err
	}
	err = p.expect(p.lexExpr, end)
	if err != nil {
		return nil, err
	}
	return &Inner{op, args}, nil
}

// e-list ::= expr ("," expr)*
// the base expression appears as the first entry in the array returned
func (p *parser) parseExprList(base Expr) ([]Expr, error) {
	list := []Expr{}
	if base != nil {
		list = append(list, base)
	}
	for {
		e, err := p.expr(0)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if p.lookExpr() != ',' {
			return list, nil
		}
		p.advanceExpr()
	}
}

// primary1 ::= identifier | integer | real | string | "/" re "/" | "@" | "$" | "(" expr ")" | "[" e-list "]" | "-" primary1 | "!" primary1
func (p *parser) primary1() (Expr, error) {
	lx := p.lexExpr()
	if lx.err != nil {
		return nil, lx.err
	}
	switch lx.tok {
	case tokError:
		return nil, lx.err
	case '-':
		return p.unary(OpNeg)
	case '!':
		return p.unary(OpNot)
	case tokID:
		return &NameLeaf{OpID, lx.s()}, nil
	case tokInt:
		return &IntLeaf{OpInt, lx.i()}, nil
	case tokReal:
		return &FloatLeaf{OpReal, lx.f()}, nil
	case tokString:
		return &StringLeaf{OpString, lx.s()}, nil
	case '/':
		lx = p.lexRegExp('/')
		if lx.err != nil {
			return nil, lx.err
		}
		return &RegexpLeaf{OpRE, lx.s()}, nil
	case '@':
		return &NameLeaf{OpCurrent, "@"}, nil
	case '$':
		return &NameLeaf{OpRoot, "$"}, nil
	case '(':
		e, err := p.parseScriptExpr()
		if err != nil {
			return nil, err
		}
		err = p.expect(p.lexExpr, ')')
		if err != nil {
			return nil, err
		}
		return e, nil
	case '[':
		// array-literal
		return p.application(OpArray, ']', nil)
	default:
		return nil, fmt.Errorf("unexpected token %v in expression term", lx.tok)
	}
}

// convert tokens to expression operators
func tok2op(t token) Op {
	switch t {
	case '*':
		return OpMul
	case '+':
		return OpAdd
	case '-':
		return OpSub
	case '/':
		return OpDiv
	case '%':
		return OpMod
	case tokEQ:
		return OpEQ
	case tokNE:
		return OpNE
	case '<':
		return OpLT
	case tokLE:
		return OpLE
	case tokGE:
		return OpGE
	case '>':
		return OpGT
	case tokAnd:
		return OpAnd
	case tokOr:
		return OpOr
	case '~', tokMatch:
		return OpMatch
	case tokIn:
		return OpIn
	case tokNin:
		return OpNin
	default:
		return OpError
	}
}
