package JSONPath

import (
	"fmt"
	"strings"
)

// Path is a sequence of Steps, following the grammar.
// The first step is always OpRoot.
type Path []*Step

// Step represents a single step in the path.
type Step struct {
	Op   Op	// Op is the action to take at this step. Not all Ops are valid Steps (eg, expression operators).
	Args []Val	// Zero or more arguments to the operation (eg, integer and string values, an identifier, a Slice or a filter or other Expr).
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

// Slice is a Val that represents a JavaScript slice with [start: end: stride], where any of them might be optional.
// It appears as an operand in an OpUnion.
type Slice struct {
	Start  Val	// optional starting offset
	End    Val	// optional end offset (exclusive)
	Stride Val	// optional value selecting every n array elements.
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
	return sb.String()
}
