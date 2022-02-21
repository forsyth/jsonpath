package paths

import (
	"fmt"
	"strings"
)

// Path is a sequence of Steps, starting from "$" (the document root),  following the grammar.
// The initial "$" has no explicit representation in the Path: it's the starting point.
type Path []*Step

// Step represents a single step in the path: an operation with zero or more parameters, each represented by a Val,
// which is either a constant (signed integer, string, or member name) or an Expr to be evaluated.
type Step struct {
	Op   Op    // Op is the action to take at this step. Not all Ops are valid Steps (eg, expression operators).
	Args []Val // Zero or more arguments to the operation (eg, integer and string values, an identifier, a Slice or a filter or other Expr).
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
