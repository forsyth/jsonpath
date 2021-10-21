package JSONPath

import (
	"fmt"
	"strings"
)

// Val is an int64, float64, string literal, name, bool?, *Slice or Expr as a value (see IntVal etc below), or nil as a missing value.
// It represents a parameter to a Path Step, or a value at the leaf of an expression tree.
// Originally, it was also the memory type for the Program machine, but because that also includes JSON trees,
// it was clearer just to use the JSON type (ie, interface{}) directly, hence the addition of Value().
// Other files in this package add their own Val variants.
type Val interface {
	String() string	// String returns a text representation, mainly for tracing and testing.
//	Value() JSON	// Value returns a suitable internal representation for use by the machine.
}

// IntVal represents an integer value, satisfying Val.
type IntVal int64

// V returns the underlying integer.
func (v IntVal) V() int64 {
	return int64(v)
}

func (v IntVal) String() string {
	return fmt.Sprint(int64(v))
}

// NameVal represents a key or JSON member name as a value, satisfying Val.
type NameVal string

func (v NameVal) String() string {
	return string(v)
}

// S returns the name as a plain string.
func (v NameVal) S() string {
	return string(v)
}

// StringVal represents a string value, satisfying Val.
type StringVal string

func (v StringVal) String() string {
	return fmt.Sprintf("%q", string(v))
}

// S returns the string value unwrapped.
func (v StringVal) S() string {
	return string(v)
}

// Slice represents a JavaScript slice with [start: end: stride], where any of them might be optional (nil).
// *Slice satisfies Val.
type Slice struct {
	Start  Val // optional starting offset
	End    Val // optional end offset (exclusive)
	Stride Val // optional value selecting every n array elements.
}

func (slice *Slice) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	if slice.Start != nil {
		sb.WriteString(slice.Start.String())
	}
	sb.WriteByte(':')
	if slice.End != nil {
		sb.WriteString(slice.End.String())
	}
	if slice.Stride != nil {
		sb.WriteByte(':')
		sb.WriteString(slice.Stride.String())
	}
	sb.WriteByte(']')
	return sb.String()
}
