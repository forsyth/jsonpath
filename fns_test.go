package JSONPath

import (
//	"encoding/json"
	"fmt"
	"testing"
)

type funcTest struct {
	expr	string	// expression with a single call or array of calls
	expect	string	// expected result
}

var funcTests = []funcTest {
	{"abs(-1.5)", "1.5"},
	{"abs(20)", "20"},
	{"avg([1, 2, 3, 4, 5])", "3"},
	{"ceil(1.5)", "2"},
}

func TestFunctions(t *testing.T) {
	for i, ft := range funcTests {
		expr, err := ParseScriptExpression(ft.expr)
		if err != nil {
			t.Fatalf("function test %d: parse %q: %s", i, ft.expr, err)
		}
		switch expr.Opcode() {
		case OpCall:
			_ = call(t, i, &ft, expr.(*Inner).kids)
		case OpArray:
			a := expr.(*Inner)
			result := []JSON{}
			for _, v := range a.kids {
				if v.Opcode() != OpCall {
					t.Fatalf("function test %d: unexpected operator in array: %#v", i, v.Opcode())
				}
				result = append(result, call(t, i, &ft, v.(*Inner).kids))
			}
			_ = result
		default:
			t.Fatalf("function test %d: unexpected root op: %#v", i, expr.Opcode())
		}
	}
}

func call(t *testing.T, tno int, ft *funcTest, kids []Expr) JSON {
	if len(kids) < 2 {
		t.Fatalf("function test %d: %q: too few children", tno, ft.expr)
	}
	if nm, ok := kids[0].(*NameLeaf); !ok {
		t.Fatalf("function test %d: %q: got %#v, expected OpID", tno, ft.expr, kids[0].Opcode())
	} else {
		id := nm.Name
		fn := functions[id]
		if fn.fn == nil {
			t.Fatalf("function test %d: %q: unknown function %s", tno, ft.expr, id)
		}
		if fn.na != AnyNumber && len(kids) != fn.na + 1 {
			t.Fatalf("function test %d: %q: wrong arg count (need %d)", tno, ft.expr, fn.na)
		}
		args, err := collect(kids[1:], []JSON{})
		if err != nil {
			t.Fatalf("function test %d: %q: %s", tno, ft.expr, err)
		}
		result := fn.fn(args)
		fmt.Printf("%s: %#v\n", id, result)
	}
	return nil
}

func collect(kids []Expr, args []JSON) ([]JSON, error) {
	for _, k := range kids {
		switch t := k.(type) {
		case *IntLeaf:
			args = append(args, t.Val)
		case *FloatLeaf:
fmt.Printf("float: %#v\n", t.Val)
			args = append(args, t.Val)
		case *StringLeaf:
			args = append(args, t.Val)
		case *Inner:
			switch t.Op {
			case OpNeg:
				switch r := t.kids[0].(type) {
				case *IntLeaf:
					args = append(args, -r.Val)
				case *FloatLeaf:
					args = append(args, -r.Val)
				default:
					return nil, fmt.Errorf("unexpected op %#v under OpNeg", t.Op)
				}
			case OpArray:
				els, err := collect(t.kids, []JSON{})
				if err != nil {
					return nil, err
				}
				args = append(args, els)
			default:
				return nil, fmt.Errorf("unexpected op %#v in argument", t.Op)
			}
		default:
			return nil, fmt.Errorf("unexpected op %#v in argument list", k.Opcode())
		}
	}
	return args, nil
}
