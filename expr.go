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
	lexer *lexer // source of tokens
}

func (p *parser) lexExpr(withRE bool) lexeme {
	return p.lexer.lexExpr(withRE)
}

func (p *parser) unget(lex lexeme) {
	p.lexer.unget(lex)
}

func (p *parser) lookExpr(withRE bool) token {
	return p.lexer.look(p.lexer.lexExpr(withRE))
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
		c := p.lookExpr(false)
		switch c {
		case '-':
			return p.unary(OpNeg, pri)
		case '!':
			return p.unary(OpNot, pri)
		}
		pri++ // primary
	}
	e, err := p.expr(pri + 1)
	if err != nil {
		return nil, err
	}
	// associate operators at current priority level
	for isOpIn(tok2op(p.lookExpr(false)), prectab[pri]) {
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

// unary applies a unary operator to a following expression
func (p *parser) unary(op Op, pri int) (Expr, error) {
	p.lexExpr(false)
	arg, err := p.expr(pri + 1)
	if err != nil {
		return nil, err
	}
	return &Inner{op, []Expr{arg}}, nil
}

func (p *parser) primary() (Expr, error) {
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
