package JSONPath

// Elements of an expression tree (Expr)

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

// IntLeaf represents an integer in an Expr tree.
type IntLeaf struct {
	Op
	IntVal
}

// NameLeaf represents a user-defined name (OpId), "@" (OpCurrent) and "$" (OpRoot) in an Expr tree.
type NameLeaf struct {
	Op
	NameVal
}

// FloatLeaf represents a floating-point number in an Expr tree.
type FloatLeaf struct {
	Op
	FloatVal
}

// StringLeaf represents the value of a single- or double-quoted string in an Expr tree.
type StringLeaf struct {
	Op
	StringVal
}

// RegexpLeaf represents the text of a regular expression an Expr tree.
type RegexpLeaf struct {
	Op
	Pattern string
}

// Expr represents an arbitrary expression tree; it can be converted to one of the ...Leaf types or Inner, depending on Op,
// or using a type switch on an Expr value.
type Expr interface {
	// Opcode gives the node's operator, which determines the detailed structure.
	// Useful to avoid .(type) for simple routing.
	Opcode() Op

	// IsLeaf is Op().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool
}
