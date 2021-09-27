package JSONPath

// Elements of an expression tree (Expr)

// Inner represents an interior operation with one or more operands.
type Inner struct {
	op   Op
	kids []Expr
}

// Isleaf returns false.
func (i *Inner) IsLeaf() bool { return false }

// Op returns the node's operator.
func (i *Inner) Op() Op { return i.op }

// Kid returns child c (index c) or nil if there isn't one.
func (i *Inner) Kid(c int) Expr {
	if c >= len(i.kids) {
		return nil
	}
	return i.kids[c]
}

// IntLeaf represents an integer.
type IntLeaf struct {
	op  Op
	val int64
}

// IsLeaf returns true.
func (l *IntLeaf) IsLeaf() bool { return true }

// Op returns the node's operator.
func (l *IntLeaf) Op() Op { return l.op }

// NameLeaf represents a user-defined name (OpId), "@" (OpCurrent) and "$" (OpRoot).
type NameLeaf struct {
	op   Op
	name string
}

// IsLeaf returns true.
func (l *NameLeaf) IsLeaf() bool { return true }

// Op returns the node's operator.
func (l *NameLeaf) Op() Op { return l.op }

// FloatLeaf represents a floating-point number.
type FloatLeaf struct {
	op  Op
	val float64
}

// IsLeaf returns true.
func (l *FloatLeaf) IsLeaf() bool { return true }

// Op returns the node's operator.
func (l *FloatLeaf) Op() Op { return l.op }

// StringLeaf represents the value of a single- or double-quoted string.
type StringLeaf struct {
	op  Op
	val string
}

// IsLeaf returns true.
func (l *StringLeaf) IsLeaf() bool { return true }

// Op returns the node's operator.
func (l *StringLeaf) Op() Op { return l.op }

// RegexpLeaf represents the value of a single- or double-quoted string.
type RegexpLeaf struct {
	op  Op
	val string
}

// IsLeaf returns true.
func (l *RegexpLeaf) IsLeaf() bool { return true }

// Op returns the node's operator.
func (l *RegexpLeaf) Op() Op { return l.op }

// Expr represents an arbitrary expression tree; it can be converted to one of the ...Leaf types or Inner, depending on Op.
type Expr interface {
	// Op gives the node's operator, which determines the detailed structure.
	Op() Op

	// IsLeaf is Op().IsLeaf() for convenience. If !IsLeaf(), it's an Inner operator.
	IsLeaf() bool
}
