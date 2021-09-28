package JSONPath

// Elements of an expression tree (Expr)

// Inner represents an interior operation with one or more operands.
type Inner struct {
	Op
	kids []Expr
}

// Kid returns child c (index c) or nil if there isn't one.
func (i *Inner) Kid(c int) Expr {
	if c >= len(i.kids) {
		return nil
	}
	return i.kids[c]
}

// IntLeaf represents an integer.
type IntLeaf struct {
	Op
	Val int64
}

// NameLeaf represents a user-defined name (OpId), "@" (OpCurrent) and "$" (OpRoot).
type NameLeaf struct {
	Op
	Name string
}

// FloatLeaf represents a floating-point number.
type FloatLeaf struct {
	Op
	Val float64
}

// StringLeaf represents the value of a single- or double-quoted string.
type StringLeaf struct {
	Op
	Val string
}

// RegexpLeaf represents the value of a single- or double-quoted string.
type RegexpLeaf struct {
	Op
	Pattern string
}

// Expr represents an arbitrary expression tree; it can be converted to one of the ...Leaf types or Inner, depending on Op.
type Expr interface {
	// Opcode gives the node's operator, which determines the detailed structure.
	// Useful to avoid .(type) for simple routing.
	Opcode() Op

	// IsLeaf is Op().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool
}
