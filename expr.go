package JSONPath

// Elements of an expression tree (Expr)

import "fmt"

// Expr represents an arbitrary expression tree; it can be converted to one of the ...Leaf types or Inner, depending on Opcode,
// or using a type switch on an Expr value. Note that Expr satisfies Val, and can appear directly as a value
// or argument in a Path.
type Expr interface {
	// Opcode gives the node's operator, which determines the detailed structure.
	// Useful to avoid .(type) for simple routing.
	Opcode() Op

	// IsLeaf is Op().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool

	String() string
}

// Inner represents an interior operation with one or more operands.
type Inner struct {
	Op
	kids []Expr
}

// Kid returns child c (index c) and true, or nil and false if the child doesn't exist.
func (i *Inner) Kid(c int) (Expr, bool) {
	if c >= len(i.kids) {
		return nil, false
	}
	return i.kids[c], true
}

func (i *Inner) String() string {
	return fmt.Sprintf("%#v", i)
}

// IntLeaf represents an integer in an Expr tree.
type IntLeaf struct {
	Op
	Val	IntVal
}

func (l *IntLeaf) String() string {
	return l.Val.String()
}

// FloatLeaf represents a floating-point number in an Expr tree.
type FloatLeaf struct {
	Op
	Val FloatVal
}

func (l *FloatLeaf) String() string {
	return l.Val.String()
}

// StringLeaf represents the value of a single- or double-quoted string in an Expr tree.
type StringLeaf struct {
	Op
	Val StringVal
}

func (l *StringLeaf) String() string {
	return l.Val.String()
}

// NameLeaf represents a user-defined name (OpId), "@" (OpCurrent) and "$" (OpRoot) in an Expr tree.
type NameLeaf struct {
	Op
	Name string
}

func (l *NameLeaf) String() string {
	return l.Name
}

// RegexpLeaf represents the text of a regular expression an Expr tree.
type RegexpLeaf struct {
	Op
	Pattern string
}

func (l *RegexpLeaf) String() string {
	return fmt.Sprintf("Regexp(%q)", l.Pattern)
}
