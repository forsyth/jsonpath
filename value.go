package JSONPath

import "fmt"

// Val is an int64, float64, string (literal or identifier), bool, *Slice or Expr as a value, or nil as a missing value.
type Val interface{
	String() string
}

// IntVal represents an integer Value
type IntVal int64

func (v IntVal) String() string {
	return fmt.Sprint(int64(v))
}

// FloatVal represents a floating-point Value
type FloatVal float64

func (v FloatVal) String() string {
	return fmt.Sprint(float64(v))
}

// NameVal represents a key or JSON member name as a Value
type NameVal string

func (v NameVal) String() string {
	return string(v)
}

// S returns the name as a plain string.
func (v NameVal) S() string {
	return string(v)
}

// StringVal represents a string Value
type StringVal string

func (v StringVal) String() string {
	return fmt.Sprintf("%q", string(v))
}

// S returns the string value unwrapped.
func (v StringVal) S() string {
	return string(v)
}
