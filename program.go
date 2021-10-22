package JSONPath

import (
	"fmt"
	"regexp"
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
	if o&smallFlag != 0 {
		panic("index cannot be small int")
	}
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

// pc extracts a pc value stored as a small int.
func (o order) pc() int {
	if o&smallFlag == 0 {
		panic("pc must be small int")
	}
	return int(o.smallInt())
}

// floatVal extends Val to include floating-point in a Program.
type floatVal float64

func (f floatVal) String() string {
	return fmt.Sprint(float64(f))
}

// Value satisfies Valuer, boxing an ordinary floating-point number.
func (f floatVal) Value() JSON {
	return float64(f)
}

func (f floatVal) F() float64 {
	return float64(f)
}

// regexpVal extends Val to include compiled regular expressions in a Program.
type regexpVal struct {
	*regexp.Regexp
}

// Value satisfies Valuer, boxing the compiled regular expression.
func (r regexpVal) Value() JSON {
	return r.Regexp
}

func (r regexpVal) String() string {
	return fmt.Sprintf("%q", r.Regexp.String())
}

// Program is the compiled form of a Path and associated expressions.
// It is a program for a simple stack machine, although the details are hidden.
type Program struct {
	vals   []Val   // unique data values, indexed by an order's index value
	orders []order // program text
}

// asm adds an instruction to the program and returns its pc.
func (p *Program) asm(o order) int {
	p.orders = append(p.orders, o)
	return p.size() - 1
}

// patch replaces the order at the given pc.
func (p *Program) patch(pc int, o order) {
	p.orders[pc] = o
}

// data adds a data value to the program and returns its index.
func (p *Program) data(val Val) uint32 {
	o := uint32(len(p.vals))
	p.vals = append(p.vals, val)
	return o
}

// value returns the data value at the given index.
func (p *Program) value(index uint32) Val {
	return p.vals[index]
}

// size returns the current size of the program in orders,
// which acts as current pc value during assembly.
func (p *Program) size() int {
	return len(p.orders)
}
