package jsonpath

import (
	"errors"
	"fmt"
)

var (
	ErrTooManyVals = errors.New("program has too many values")
)

// builder is the state when building a program
type builder struct {
	vals map[Val]uint32 // map value to its index in prog.vals
	prog *Program
}

// Compile compiles a Path into a Program for a small abstract machine that evaluates paths and expressions.
func (path Path) Compile() (*Program, error) {
	prog := &Program{}
	b := &builder{vals: make(map[Val]uint32), prog: prog}
	for _, step := range path {
		if step.Op.IsLeaf() && step.Op.HasVal() {
			// leaf carries a value index
			err := b.codeVal(step.Op, step.Args[0])
			if err != nil {
				return nil, err
			}
			continue
		}
		switch step.Op {
		case OpNestMember, OpNestFilter, OpNestSelect, OpNestUnion, OpNestWild:
			err := b.codeLoop(step, OpNest)
			if err != nil {
				return nil, err
			}
		case OpFilter:
			err := b.codeLoop(step, OpFor)
			if err != nil {
				return nil, err
			}
		default:
			// general case
			_, err := b.codeStep(step)
			if err != nil {
				return nil, err
			}
		}
	}
	return prog, nil
}

func (b *builder) codeLoop(step *Step, intro Op) error {
	prog := b.prog
	fpc := prog.asm(mkSmall(intro, 0))
	lpc, err := b.codeStep(step)
	if err != nil {
		return err
	}
	prog.asm(mkSmall(OpRep, lpc))
	prog.patch(fpc, mkSmall(intro, prog.size()))
	return nil
}

func (b *builder) codeStep(step *Step) (int, error) {
	prog := b.prog
	pc := prog.size()
	err := b.codeArgs(step.Args)
	prog.asm(mkSmall(step.Op, len(step.Args)))
	return pc, err
}

func (b *builder) codeArgs(args []Val) error {
	for _, arg := range args {
		err := b.codeVal(valOp(arg), arg)
		if err != nil {
			return err
		}
	}
	return nil
}

// valOp returns the best op for the value.
func valOp(arg Val) Op {
	switch arg.(type) {
	case NameVal:
		return OpID
	case IntVal:
		return OpInt
	case StringVal:
		return OpString
	case *Slice:
		return OpBounds
	case Expr:
		return OpExp
	default:
		panic(fmt.Sprintf("unexpected valOp: %#v", arg))
	}
}

func (b *builder) codeVal(op Op, val Val) error {
	if expr, ok := val.(Expr); ok {
		return b.codeExpr(expr)
	}
	if !op.HasVal() || val == nil {
		return b.codeOp(op, nil)
	}
	return b.codeOp(op, val)
}

func (b *builder) codeExpr(expr Expr) error {
	if expr == nil {
		panic("unexpected nil expr")
	}
	if expr.IsLeaf() {
		return b.codeLeaf(expr)
	}
	t := expr.(*Inner)
	for _, k := range t.kids {
		err := b.codeExpr(k)
		if err != nil {
			return err
		}
	}
	b.prog.asm(mkSmall(t.Op, len(t.kids)))
	return nil
}

func (b *builder) codeLeaf(expr Expr) error {
	op := expr.Opcode()
	if !op.HasVal() {
		return b.codeOp(op, nil)
	}
	switch l := expr.(type) {
	case *IntLeaf:
		if isSmallInt(l.Val) {
			// skip conversion via IntVal
			b.prog.asm(mkSmall(op, int(l.Val)))
			return nil
		}
		return b.codeOp(op, IntVal(l.Val))
	case *FloatLeaf:
		return b.codeOp(op, floatVal(l.Val))
	case *StringLeaf:
		return b.codeOp(op, StringVal(l.Val))
	case *NameLeaf:
		return b.codeOp(op, NameVal(l.Name))
	case *RegexpLeaf:
		return b.codeOp(op, regexpVal{l.Prog})
	case *BoolLeaf:
		var v int
		if l.Val {
			v = 1
		}
		b.prog.asm(mkSmall(op, v))
		return nil
	case *NullLeaf:
		return b.codeOp(op, nil)
	default:
		panic("unexpected Leaf op: " + op.GoString())
	}
}

func (b *builder) codeOp(op Op, val Val) error {
	if val == nil {
		b.prog.asm(mkSmall(op, 0))
		return nil
	}
	if v, ok := val.(IntVal); ok {
		n := v.V()
		if isSmallInt(n) {
			b.prog.asm(mkSmall(op, int(n)))
			return nil
		}
	}
	index, err := b.mkVal(val)
	if err != nil {
		return err
	}
	b.prog.asm(mkOrder(op, index))
	return nil
}

// mkVal assigns and returns an index for val, for the index field of an order.
func (b *builder) mkVal(val Val) (uint32, error) {
	i, ok := b.vals[val]
	if ok {
		return i, nil
	}
	o := b.prog.data(val)
	if o >= indexTop {
		// more then 8m
		return 0, ErrTooManyVals
	}
	b.vals[val] = o
	return o, nil
}
