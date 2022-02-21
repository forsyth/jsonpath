package mach

import (
	"errors"
	"fmt"

	"github.com/forsyth/jsonpath/paths"
)

var (
	ErrTooManyVals = errors.New("program has too many values")
)

// builder is the state when building a program
type builder struct {
	vals map[paths.Val]uint32 // map value to its index in prog.vals
	prog *Program
}

// Compile compiles a Path into a Program for a small abstract machine that evaluates paths and expressions.
func Compile(path paths.Path) (*Program, error) {
	prog := &Program{}
	b := &builder{vals: make(map[paths.Val]uint32), prog: prog}
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
		case paths.OpNestMember, paths.OpNestFilter, paths.OpNestSelect, paths.OpNestUnion, paths.OpNestWild:
			err := b.codeLoop(step, paths.OpNest)
			if err != nil {
				return nil, err
			}
		case paths.OpFilter:
			err := b.codeLoop(step, paths.OpFor)
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

func (b *builder) codeLoop(step *paths.Step, intro paths.Op) error {
	prog := b.prog
	fpc := prog.asm(mkSmall(intro, 0))
	lpc, err := b.codeStep(step)
	if err != nil {
		return err
	}
	prog.asm(mkSmall(paths.OpRep, lpc))
	prog.patch(fpc, mkSmall(intro, prog.size()))
	return nil
}

func (b *builder) codeStep(step *paths.Step) (int, error) {
	prog := b.prog
	pc := prog.size()
	err := b.codeArgs(step.Args)
	prog.asm(mkSmall(step.Op, len(step.Args)))
	return pc, err
}

func (b *builder) codeArgs(args []paths.Val) error {
	for _, arg := range args {
		err := b.codeVal(valOp(arg), arg)
		if err != nil {
			return err
		}
	}
	return nil
}

// valOp returns the best op for the value.
func valOp(arg paths.Val) paths.Op {
	switch arg.(type) {
	case paths.NameVal:
		return paths.OpID
	case paths.IntVal:
		return paths.OpInt
	case paths.StringVal:
		return paths.OpString
	case *paths.Slice:
		return paths.OpBounds
	case paths.Expr:
		return paths.OpExp
	default:
		panic(fmt.Sprintf("unexpected valOp: %#v", arg))
	}
}

func (b *builder) codeVal(op paths.Op, val paths.Val) error {
	if expr, ok := val.(paths.Expr); ok {
		return b.codeExpr(expr)
	}
	if !op.HasVal() || val == nil {
		return b.codeOp(op, nil)
	}
	return b.codeOp(op, val)
}

func (b *builder) codeExpr(expr paths.Expr) error {
	if expr == nil {
		panic("unexpected nil expr")
	}
	if expr.IsLeaf() {
		return b.codeLeaf(expr)
	}
	t := expr.(*paths.Inner)
	for _, k := range t.Kids {
		err := b.codeExpr(k)
		if err != nil {
			return err
		}
	}
	b.prog.asm(mkSmall(t.Op, len(t.Kids)))
	return nil
}

func (b *builder) codeLeaf(expr paths.Expr) error {
	op := expr.Opcode()
	if !op.HasVal() {
		return b.codeOp(op, nil)
	}
	switch l := expr.(type) {
	case *paths.IntLeaf:
		if isSmallInt(l.Val) {
			// skip conversion via IntVal
			b.prog.asm(mkSmall(op, int(l.Val)))
			return nil
		}
		return b.codeOp(op, paths.IntVal(l.Val))
	case *paths.FloatLeaf:
		return b.codeOp(op, floatVal(l.Val))
	case *paths.StringLeaf:
		return b.codeOp(op, paths.StringVal(l.Val))
	case *paths.NameLeaf:
		return b.codeOp(op, paths.NameVal(l.Name))
	case *paths.RegexpLeaf:
		return b.codeOp(op, regexpVal{l.Prog})
	case *paths.BoolLeaf:
		var v int
		if l.Val {
			v = 1
		}
		b.prog.asm(mkSmall(op, v))
		return nil
	case *paths.NullLeaf:
		return b.codeOp(op, nil)
	default:
		panic("unexpected Leaf op: " + op.GoString())
	}
}

func (b *builder) codeOp(op paths.Op, val paths.Val) error {
	if val == nil {
		b.prog.asm(mkSmall(op, 0))
		return nil
	}
	if v, ok := val.(paths.IntVal); ok {
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
func (b *builder) mkVal(val paths.Val) (uint32, error) {
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
