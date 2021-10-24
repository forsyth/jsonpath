package JSONPath

import (
	"errors"
	"fmt"
	"testing"
)

type funcTest struct {
	expr   string // expression with a single call or array of calls
	expect string // expected result
}

var funcTests = []funcTest{
	{"abs(-1.5)", "1.5"},
	{"abs(20)", "20"},
	{"avg([1, 2, 3, 4, 5])", "3"},
	{"ceil(1.5)", "2"},
	{"ceil(2)", "2"},
	{"ceil(-1.75)", "-1"},
	{"contains('subject stringy', 'ject')", "true"},
	{"contains('subject stringy', 'queen')", "false"},
	{"contains([1, 2, 3.5, 'hat', 6], 'hat')", "true"},
	{"ends_with('and another thing', 'other thing')", "true"},
	{"ends_with('christmas', 'dinner')", "false"},
	{"floor(-1.75)", "-2"},
	{"floor(1.75)", "1"},
	{"floor(1)", "1"},
	// without object literals, keys can't yet be tested this way: needs to be part of path test
	{"length('hello, sailor')", "13"},
	{"length([1, 2, 3, 4, 5])", "5"},
	{"length([])", "0"},
	{"length(2.5)", "null"},
	{"max(1, 2, 3, 5, 4)", "5"},
	{"max([-1, -2, -3, 5, 4])", "5"},
	{"min(5, 3, 1, -1, 0)", "-1"},
	{"min([5, 3.5, 1, -1.5, 0])", "-1.5"},
	{"prod([])", "null"},
	{"prod([1, 2, 3, 4, 5])", "120"},
	{"starts_with('christmas', 'chr')", "true"},
	{"starts_with('christmas', 'all hallows')", "false"},
	{"sum([])", "0"},
	{"sum([1, 2, 3, 4, 5.55])", "15.55"},
	{"to_number(1.75)", "1.75"},
	//	{"to_number('apple')", ""},
	{"to_number('1.75e5')", "175000"},
	{"tokenize('J. R. R. Tolkien', '[[:space:]]+')", "[\"J.\",\"R.\",\"R.\",\"Tolkien\"]"},
}

func TestFunctions(t *testing.T) {
	for i, ft := range funcTests {
		expr, err := ParseScriptExpression(ft.expr)
		if err != nil {
			t.Fatalf("function test %d: parse %q: %s", i, ft.expr, err)
		}
		var got string
		switch expr.Opcode() {
		case OpCall:
			ret, err := testcall(expr.(*Inner).kids)
			if err != nil {
				t.Fatalf("function test %d: %q: %s", i, ft.expr, err)
			}
			got = jsonString(ret)
		case OpArray:
			a := expr.(*Inner)
			array := []JSON{}
			for _, v := range a.kids {
				if v.Opcode() != OpCall {
					t.Fatalf("function test %d: unexpected operator in array: %#v", i, v.Opcode())
				}
				ret, err := testcall(v.(*Inner).kids)
				if err != nil {
					t.Fatalf("function test %d: %q: %s", i, ft.expr, err)
				}
				array = append(array, ret)
			}
			got = jsonString(array)
		default:
			t.Fatalf("function test %d: unexpected root op: %#v", i, expr.Opcode())
		}
		fmt.Printf("%s: %#v\n", ft.expr, got)
		if got != ft.expect {
			t.Errorf("function test %d: %q: got (%s) expected (%s)", i, ft.expr, got, ft.expect)
		}
	}
}

func testcall(kids []Expr) (JSON, error) {
	if len(kids) == 0 {
		return nil, errors.New("no identifier child in call")
	}
	nm, ok := kids[0].(*NameLeaf)
	if !ok {
		return nil, fmt.Errorf("got %#v, expected OpID", kids[0].Opcode())
	}
	args, err := collect(kids[1:], []JSON{})
	if err != nil {
		return nil, fmt.Errorf("call to %s: %w", nm.Name, err)
	}
	return call(nm.Name, args)
}

func collect(kids []Expr, args []JSON) ([]JSON, error) {
	for _, k := range kids {
		switch t := k.(type) {
		case *IntLeaf:
			args = append(args, t.Val)
		case *FloatLeaf:
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
