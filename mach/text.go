package mach

import (
	"fmt"
	"strings"

	"github.com/forsyth/jsonpath/paths"
)

// String returns a representation of the text of the abstract machine program.
func (prog *Program) String() string {
	var sb strings.Builder
	for i, val := range prog.vals {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(val.String())
	}
	if len(prog.vals) > 0 {
		sb.WriteByte(' ')
	}
	for i, ord := range prog.orders {
		if i > 0 {
			sb.WriteByte(' ')
		}
		op := ord.op()
		sb.WriteString(trimOp(op))
		if op.IsLeaf() {
			if !op.HasVal() {
				continue
			}
			if ord.isSmallInt() {
				sb.WriteByte('(')
				sb.WriteString(fmt.Sprint(ord.smallInt()))
				sb.WriteByte(')')
			} else {
				sb.WriteByte('[')
				sb.WriteString(fmt.Sprint(ord.index()))
				sb.WriteByte(']')
			}
			continue
		}
		if ord.isSmallInt() && ord.smallInt() != 0 {
			sb.WriteByte('.')
			sb.WriteString(fmt.Sprint(ord.smallInt()))
		}
	}
	return sb.String()
}

func trimOp(op paths.Op) string {
	name := fmt.Sprintf("%#v", op)
	if name == "" {
		panic("unknown op in opNames")
	}
	return name[2:]
}
