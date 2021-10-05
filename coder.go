package JSONPath

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTooManyVals = errors.New("program has too many values")
)

// order holds a single order for the path machine.
// The low order 8 bits hold an Op.
// smallFlag is the next bit.
// The next 23 bits is the index field, holding an integer value if smallFlag is set,
// or a Val table index if smallFlag is zero. OpCall and OpArray use the value as
// an operand count, since those are the only operations with zero or more operands.
type order uint32

const (
	opSize            = 8                    // bits in the op field
	opMask            = (1 << opSize) - 1    // mask selecting the op field
	smallFlag   order = 1 << opSize          // the index field contains an integer not an index
	indexSize         = 23                   // bits in the index field
	indexOffset       = opSize + 1           // beyond op field and smallFlag
	indexTop          = 1 << (indexSize - 1) // leftmost bit in the index field, once extracted
	indexMask         = (1 << indexSize) - 1 // just the index field, once extracted
)

// mkOrder returns an order with an op and index field.
func mkOrder(op Op, index uint32) order {
	return order(op&opMask) | order(index<<indexOffset)
}

// mkSmall returns an order with an op and signed integer.
func mkSmall(op Op, val int) order {
	return order(op&opMask) | order((val&indexMask)<<indexOffset) | smallFlag
}

// op returns the operation part of an order.
func (o order) op() Op {
	return Op(o & opMask)
}

// isSmallInt is true if the order contains a small integer, not its index
func (o order) isSmallInt() bool {
	return o&smallFlag != 0
}

// index returns the index field as a Val table index.
// smallFlag must be zero (or the index will cause a panic if used).
func (o order) index() uint32 {
	return uint32(o) >> indexOffset
}

// smallInt extracts a signed integer from the index field.
func (o order) smallInt() int64 {
	// the arithmetic shift sign-extends the integer.
	f := int32(o) >> indexOffset
	return int64(f)
}

// isSmallInt returns true if signed integer i can be encoded in the index field.
func isSmallInt(i int64) bool {
	return i >= -indexTop && i <= indexTop-1
}

// Program is the compiled form of a Path and associated expressions.
// It is a program for a simple stack machine, although the details are hidden.
type Program struct {
	vals   []Val   // unique data values, indexed by MachOp's index value
	orders []order // program text
}

// asm adds an instruction to the program.
func (p *Program) asm(o order) {
	p.orders = append(p.orders, o)
}

// data adds a data value to the program and returns its index.
func (p *Program) data(val Val) uint32 {
	o := uint32(len(p.vals))
	p.vals = append(p.vals, val)
	return o
}

// builder is the state when building a program
type builder struct {
	vals map[Val]uint32 // map value to its index in prog.vals
	prog *Program
}

// VM is the current state of the virtual machine.
type VM struct {
	root  interface{}   // $
	json  []interface{} // current value set (@)
	stack []Val         // expression stack
}

// FloatVal extends Val to include floating-point in a Program.
type FloatVal float64

func (f FloatVal) String() string {
	return fmt.Sprint(float64(f))
}

type BuilderWriter interface {
	writeTo(*strings.Builder)
}

// CodePath converts the Path into a Program for a small virtual machine
func CodePath(path Path) (*Program, error) {
	prog := &Program{}
	b := &builder{vals: make(map[Val]uint32), prog: prog}
	for _, step := range path {
		if step.Op.IsLeaf() {
			if step.Op.HasVal() {
				// leaf carries a value index
				err := b.codeVal(step.Op, step.Args[0])
				if err != nil {
					return nil, err
				}
			} else {
				// leaf Op implies a value
				prog.asm(mkSmall(step.Op, 0))
			}
			continue
		}
		if len(step.Args) > 0 {
			for _, arg := range step.Args {
				err := b.codeVal(OpVal, arg)
				if err != nil {
					return nil, err
				}
			}
		}
		prog.asm(mkSmall(step.Op, len(step.Args)))
	}
	return prog, nil
}

func (b *builder) codeVal(op Op, val Val) error {
	if !op.HasVal() || val == nil {
		return b.codeOp(op, nil)
	}
	if expr, ok := val.(Expr); ok {
		return b.codeExpr(expr)
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
		return b.codeOp(op, FloatVal(l.Val))
	case *StringLeaf:
		return b.codeOp(op, StringVal(l.Val))
	case *NameLeaf:
		return b.codeOp(op, NameVal(l.Name))
	case *RegexpLeaf:
		// might need a RegexpVal to hold compiled version
		return b.codeOp(op, StringVal(l.Pattern))
	default:
		panic("unexpected Leaf op: " + op.GoString())
		return nil
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
