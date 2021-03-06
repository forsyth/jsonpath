package paths

// Elements of an expression tree (Expr)

import (
	"fmt"
	"regexp"
)

// Expr represents an arbitrary expression tree; it can be converted to one of the ...Leaf types or Inner, depending on Opcode,
// or using a type switch on an Expr value. Note that Expr satisfies Val, and can appear directly as a value
// or argument in a Path. Since it is not a constant value, it is not a Valuer.
type Expr interface {
	// Opcode gives the node's operator, which determines the detailed structure.
	// Useful to avoid .(type) for simple routing.
	Opcode() Op

	// IsLeaf is Opcode().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool

	String() string
}

// Inner represents an interior operation with one or more operands.
type Inner struct {
	Op
	Kids []Expr
}

// Kid returns child c (index c) and true, or nil and false if the child doesn't exist.
func (i *Inner) Kid(c int) (Expr, bool) {
	if c >= len(i.Kids) {
		return nil, false
	}
	return i.Kids[c], true
}

func (i *Inner) String() string {
	return fmt.Sprintf("%#v", i)
}

// IntLeaf represents an integer in an Expr tree.
type IntLeaf struct {
	Op
	Val int64
}

func (l *IntLeaf) String() string {
	return fmt.Sprint(l.Val)
}

// FloatLeaf represents a floating-point number in an Expr tree.
type FloatLeaf struct {
	Op
	Val float64
}

func (l *FloatLeaf) String() string {
	return fmt.Sprint(l.Val)
}

// StringLeaf represents the value of a single- or double-quoted string in an Expr tree.
type StringLeaf struct {
	Op
	Val string
}

func (l *StringLeaf) String() string {
	return fmt.Sprintf("%q", l.Val)
}

// BoolLeaf represents a Boolean in an Expr tree.
type BoolLeaf struct {
	Op
	Val bool
}

func (l *BoolLeaf) String() string {
	return fmt.Sprint(l.Val)
}

// NullLeaf represents JS "null" in an Expr tree.
type NullLeaf struct {
	Op
}

func (l *NullLeaf) String() string {
	return "null"
}

// NameLeaf represents a user-defined name (OpID), "@" (OpCurrent) and "$" (OpRoot) in an Expr tree.
type NameLeaf struct {
	Op
	Name string
}

func (l *NameLeaf) String() string {
	return l.Name
}

// RegexpLeaf represents the text of a regular expression in an Expr tree.
type RegexpLeaf struct {
	Op
	Pattern string         // Pattern is the text of the expression.
	Prog    *regexp.Regexp // Prog is the compiled version of the same.
}

func (l *RegexpLeaf) String() string {
	return fmt.Sprintf("Regexp(%q)", l.Pattern)
}
