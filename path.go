package JSONPath

import (
	"fmt"
	"strings"
)

// Path is a sequence of Steps, following the grammar.
// The first step is always OpRoot.
type Path []*Step

// Step represents a single step in the path: an operation with zero or more parameters, each represented by a Val,
// which is either a constant (signed integer, string, or member name) or an Expr to be evaluated.
type Step struct {
	Op   Op    // Op is the action to take at this step. Not all Ops are valid Steps (eg, expression operators).
	Args []Val // Zero or more arguments to the operation (eg, integer and string values, an identifier, a Slice or a filter or other Expr).
}

// Val is an int64, float64, string literal, name, bool?, *Slice or Expr as a value (see IntVal etc below), or nil as a missing value.
type Val interface {
	String() string
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

func (step *Step) String() string {
	var sb strings.Builder
	sb.WriteString(step.Op.GoString())
	if len(step.Args) > 0 {
		sb.WriteByte('(')
		for i, a := range step.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%#v", a))
		}
		sb.WriteByte(')')
	}
	return sb.String()
}
