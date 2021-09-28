package JSONPath

import "fmt"

// Val is an int64, float64, string (literal or identifier), bool, *Slice or Expr
type Val interface{}

func ValString(v Val) string {
	switch v := v.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprint(v)
	case int64:
		return fmt.Sprint(v)
	case float64:
		return fmt.Sprint(v)
	case string:
		return fmt.Sprintf("%q", v)
	case *Slice:
		return v.String()
	case Expr:
		return "(expr)" //v.String()
	default:
		panic(fmt.Sprintf("unexpected type: %#v", v))
	}
}
