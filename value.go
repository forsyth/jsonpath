package JSONPath

import "fmt"

// Val is an int64, float64, string (literal or identifier), bool, *Slice or Expr as a value, or nil as a missing value.
type Val interface{
	String() string
}

// IntVal represents an integer Value
type IntVal struct {
	Val	int64
}

// Zero is a synonym for IntVal's zero value.
var Zero = IntVal{0}

func (v IntVal) String() string {
	return fmt.Sprint(v.Val)
}

// FloatVal represents a floating-point Value
type FloatVal struct {
	Val	float64
}

func (v FloatVal) String() string {
	return fmt.Sprint(v.Val)
}

// NameVal represents a key or JSON member name as a Value
type NameVal struct {
	Name string
}

func (v NameVal) String() string {
	return v.Name
}

// StringVal represents a string Value
type StringVal struct {
	Val string
}

func (v StringVal) String() string {
	return fmt.Sprintf("%q", v.Val)
}

type ExprVal struct {
	Expr Expr
}

func (e ExprVal) String() string {
	return fmt.Sprintf("%#v", e.Expr)
}

func (e ExprVal) IsMissing() bool {
	return e.Expr == nil
}
