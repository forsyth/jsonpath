package JSONPath

import (
	//	"errors"
	"fmt"
)

// Inner represents an expression operation with one or more operands.
type Inner struct {
	op   Op
	kids []Expr
}

func (i *Inner) IsLeaf() bool { return false }
func (i *Inner) Op() Op       { return i.op }
func (i *Inner) Kid(c int) Expr {
	if c >= len(i.kids) {
		return nil
	}
	return i.kids[c]
}

// IntLeaf represents an integer.
type IntLeaf struct {
	op  Op
	val int64
}

func (l *IntLeaf) IsLeaf() bool { return true }
func (l *IntLeaf) Op() Op       { return l.op }

// NameLeaf represents a user-defined name (OpId), "@" (OpCurrent) and "$" (OpRoot).
type NameLeaf struct {
	op   Op
	name string
}

func (l *NameLeaf) IsLeaf() bool { return true }
func (l *NameLeaf) Op() Op       { return l.op }

// FloatLeaf represents a floating-point number.
type FloatLeaf struct {
	op  Op
	val float64
}

func (l *FloatLeaf) IsLeaf() bool { return true }
func (l *FloatLeaf) Op() Op       { return l.op }

// StringLeaf represents the value of a single- or double-quoted string.
type StringLeaf struct {
	op  Op
	val string
}

func (l *StringLeaf) IsLeaf() bool { return true }
func (l *StringLeaf) Op() Op       { return l.op }

// RegexpLeaf represents the value of a single- or double-quoted string.
type RegexpLeaf struct {
	op  Op
	val string
}

func (l *RegexpLeaf) IsLeaf() bool { return true }
func (l *RegexpLeaf) Op() Op       { return l.op }

// Expr represents an arbitrary expression tree; it can be cast to one of the above, depending on Op.
type Expr interface {
	// Op gives the node's operator, which determines the detailed structure.
	Op() Op

	// IsLeaf is Op().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool
}

// prectab lists the operator precedence groups, from low to high.
var prectab [][]Op = [][]Op{
	[]Op{OpOr},
	[]Op{OpAnd},
	[]Op{OpEq, OpNe},
	[]Op{OpLt, OpLe, OpGt, OpGe, OpMatch, OpIn, OpNin},
	[]Op{OpAdd, OpSub},
	[]Op{OpMul, OpDiv, OpMod},
	[]Op{OpNeg, OpNot}, // unary '-'
	//	array[] of {'|'},	// UnionExpr
}

// parser represents the state of the expression parser
type parser struct {
	*lexer // source of tokens
}

func (p *parser) lookExpr() token {
	return p.look(p.lexExpr(false))
}

func (p *parser) advanceExpr() {
	_ = p.lexExpr(false)
}

func (p *parser) parseScriptExpr() (Expr, error) {
	return p.expr(0)
}

// parse a subexpression with priority pri
func (p *parser) expr(pri int) (Expr, error) {
	if pri >= len(prectab) {
		return p.primary()
	}
	if prectab[pri][0] == OpNeg { // unary '-' or '!'
		c := p.lookExpr()
		switch c {
		case '-':
			return p.unary(OpNeg, pri)
		case '!':
			return p.unary(OpNot, pri)
		}
		//pri++ // primary
	}
	e, err := p.expr(pri + 1)
	if err != nil {
		return nil, err
	}
	// associate operators at current priority level
	for isOpIn(tok2op(p.lookExpr()), prectab[pri]) {
		lx := p.lexExpr(false)
		if lx.err != nil {
			return nil, lx.err
		}
		right, err := p.expr(pri + 1)
		if err != nil {
			return nil, err
		}
		e = &Inner{tok2op(lx.tok), []Expr{e, right}}
	}
	return e, nil
}

// unary applies a unary operator to a following primary expression
// unary ::= ("-" | "!")+ primary
func (p *parser) unary(op Op, pri int) (Expr, error) {
	p.advanceExpr()
	arg, err := p.primary()
	if err != nil {
		return nil, err
	}
	return &Inner{op, []Expr{arg}}, nil
}

// primary ::= primary1 ("(" e-list ")" | "[" e-list "]" | "." identifier)*
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
			// index (and slice?)
			p.advanceExpr()
			e, err = p.application(OpIndex, ']', e)
			if err != nil {
				return nil, err
			}
		case '.':
			// field selection
			p.advanceExpr()
			lx := p.lexExpr(false)
			if lx.err != nil {
				return nil, lx.err
			}
			if lx.tok != tokID {
				return nil, fmt.Errorf("expected identifier in '.' selection")
			}
			e = &Inner{OpSelect, []Expr{e, &NameLeaf{OpId, lx.val.(string)}}}
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
	err = p.expect(end)
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
	}
}

// primary1 ::= identifier | integer | string | "/" re "/" | "@" | "$" | "(" expr ")"
func (p *parser) primary1() (Expr, error) {
	lx := p.lexExpr(true)
	if lx.err != nil {
		return nil, lx.err
	}
	switch lx.tok {
	case tokID:
		// (), [] handled here?
		return &NameLeaf{OpId, lx.val.(string)}, nil
	case tokInt:
		return &IntLeaf{OpInt, lx.val.(int64)}, nil
	case tokString:
		return &StringLeaf{OpString, lx.val.(string)}, nil
	case tokRE:
		return &RegexpLeaf{OpRE, lx.val.(string)}, nil
	case '@':
		return &NameLeaf{OpCurrent, "@"}, nil
	case '$':
		return &NameLeaf{OpRoot, "$"}, nil
	case '(':
		e, err := p.parseScriptExpr()
		if err != nil {
			return nil, err
		}
		err = p.expect(')')
		if err != nil {
			return nil, err
		}
		return e, nil
	case '[':
		return p.application(OpArray, ']', nil)
	default:
		return nil, fmt.Errorf("unexpected token %v in expression term", lx.tok)
	}
}

func (p *parser) expect(req token) error {
	lx := p.lexExpr(false)
	if lx.err != nil {
		return lx.err
	}
	if lx.tok != req {
		return fmt.Errorf("expected %v, got %v", req, lx.tok)
	}
	return nil
}

func opPrec(t token, p []token) int {
	for j := 0; j < len(p); j++ {
		if t == p[j] {
			return j
		}
	}
	// not an operator
	return -1
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
	case tokEq:
		return OpEq
	case tokNE:
		return OpNe
	case '<':
		return OpLt
	case tokLE:
		return OpLe
	case tokGE:
		return OpGe
	case '>':
		return OpGt
	case tokAnd:
		return OpAnd
	case tokOr:
		return OpOr
	default:
		return OpError
	}
}

func isOpIn(op Op, ops []Op) bool {
	for _, v := range ops {
		if op == v {
			return true
		}
	}
	return false
}
