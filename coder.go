package JSONPath

import (
	//	"fmt"
	"strings"
)

type MachOp uint16 // machine operation

type VM struct {
	json  []interface{} // current value set
	stack []interface{} // expression stack
}

type BuilderWriter interface {
	writeTo(*strings.Builder)
}

// codePath converts the Path into a series of operators, initially in text form for development
func codePath(path Path) string {
	var sb strings.Builder
	for i, step := range path {
		if len(step.Args) > 0 {
			for _, arg := range step.Args {
				codeVal(&sb, arg)
			}
		}
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(trimOp(step.Op))
	}
	return sb.String()
}

func trimOp(op Op) string {
	name := opNames[op]
	if name == "" {
		panic("unknown op in opNames")
	}
	return name[2:]
}

func codeVal(sb *strings.Builder, val Val) {
	if expr, ok := val.(Expr); ok {
		codeExpr(sb, expr)
		return
	}
	sb.WriteByte(' ')
	if bw, ok := interface{}(val).(BuilderWriter); ok {
		bw.writeTo(sb)
	} else {
		sb.WriteString(val.String())
	}
}

func codeExpr(sb *strings.Builder, expr Expr) {
	if expr == nil {
		sb.WriteString(" nil")
		return
	}
	if expr.IsLeaf() {
		sb.WriteByte(' ')
		sb.WriteString(expr.String())
		return
	}
	t := expr.(*Inner)
	for _, k := range t.kids {
		codeExpr(sb, k)
	}
	sb.WriteByte(' ')
	sb.WriteString(trimOp(t.Op))
}
