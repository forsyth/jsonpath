package JSONPath

import (
	"errors"
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

// NameLeaf represents a user-defined name (Oid), "@" (Ocurrent) and "$" (Oroot).
type NameLeaf struct {
	op  Op
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
	[]Op{Oor},
	[]Op{Oand},
	[]Op{Oeq, One},
	[]Op{Olt, Ole, Ogt, Oge},
	[]Op{Oadd, Osub},
	[]Op{Omul, Odiv, Omod},
	[]Op{Oneg}, // unary '-'
	//	array[] of {'|'},	// UnionExpr
}

// parser represents the state of the expression parser
type parser struct {
	r    *rd
	peek bool  // unget was called
	tok  token // 1 token lookahead
	val  Val   // associated value, if any
	err  error // associated error value, if any
}

func (p *parser) get() (token, Val, error) {
	if p.peek {
		p.peek = false
		return p.tok, p.val, p.err
	}
	return lexExpr(p.r, false)
}

func (p *parser) unget(tok token, val Val, err error) {
	if p.peek {
		panic("internal error: too much lookahead")
	}
	p.peek = true
	p.tok = tok
	p.val = val
	p.err = err
}

func (p *parser) look() token {
	tok, val, err := p.get()
	p.unget(tok, val, err)
	if err != nil {
		return tokError
	}
	return tok
}

func (p *parser) parse() (Expr, error) {
	// need to check !p.peek at end
	return nil, errors.New("parse not done yet")
}

// parse a subexpression with priority pri
func (p *parser) expr(pri int) (Expr, error) {
	if pri >= len(prectab) {
		return p.primary()
	}
	if prectab[pri][0] == Oneg { // unary '-'
		if p.look() == '-' {
			p.get()
			arg, err := p.expr(pri + 1)
			if err != nil {
				return nil, err
			}
			return &Inner{Oneg, []Expr{arg}}, nil
		}
		pri++ // ???
	}
	e, err := p.expr(pri + 1)
	if err != nil {
		return nil, err
	}
	// associate operators at current priority level
	for isOpIn(tok2op(p.look()), prectab[pri]) {
		tok, _, err := p.get()
		if err != nil {
			return nil, err
		}
		right, err := p.expr(pri + 1)
		if err != nil {
			return nil, err
		}
		e = &Inner{tok2op(tok), []Expr{e, right}}
	}
	return e, nil
}

func (p *parser) primary() (Expr, error) {
	tok, val, err := p.get()
	if err != nil {
		return nil, err
	}
	switch tok {
	case tokID:
		// (), [] handled here?
		return &NameLeaf{Oid, val.(string)}, nil
	case tokInt:
		return &IntLeaf{Oint, val.(int64)}, nil
	case tokString:
		return &StringLeaf{Ostring, val.(string)}, nil
	case '@':
		return &NameLeaf{Ocurrent, "@"}, nil
	case '$':
		return &NameLeaf{Oroot, "$"}, nil
	case '(':
		e, err := parseExpr(p.r)
		if err != nil {
			return nil, err
		}
		err = p.expect(')')
		if err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unexpected token %v in expression term", tok)
	}
}

func (p *parser) expect(req token) error {
	tok, _, err := p.get()
	if err != nil {
		return err
	}
	if tok != req {
		return fmt.Errorf("expected %v, got %v", req, tok)
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
		return Omul
	case '+':
		return Oadd
	case '-':
		return Osub
	case '/':
		return Odiv
	case '%':
		return Omod
	case tokEq:
		return Oeq
	case tokNE:
		return One
	case '<':
		return Olt
	case tokLE:
		return Ole
	case tokGE:
		return Oge
	case '>':
		return Ogt
	case tokAnd:
		return Oand
	case tokOr:
		return Oor
	default:
		return Oerror
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
