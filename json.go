package jsonpath

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// JSON is a synonym for the interface{} structures returned by encoding/json,
// used as values in the JSON machine, to make it clear that's what they are.
type JSON = interface{}

// enquiry functions on JSON, sometimes easier to read than type switches

// isString returns true if v is a string.
func isString(v JSON) bool {
	_, ok := v.(string)
	return ok
}

// isInt returns true if v is an integer (int might arise from len results).
func isInt(v JSON) bool {
	switch v.(type) {
	case int, int64:
		return true
	default:
		return false
	}
}

// isFloat returns true if v is a floating-point value.
func isFloat(v JSON) bool {
	_, ok := v.(float64)
	return ok
}

// isSlice returns true if v represents slice parameters.
func isSlice(v JSON) bool {
	_, ok := v.(*Slice)
	return ok
}

// isArith returns true if v is an arithmetic type in the JS sense.
func isArith(v JSON) bool {
	switch v.(type) {
	case int, int64, float64:
		return true
	case bool:
		return true // surprise!
	default:
		return false
	}
}

// isSimple returns true if the v is "simple" (isn't a JSON object or array).
func isSimple(v JSON) bool {
	switch v.(type) {
	case bool, int, int64, float64, string:
		return true
	default:
		return false
	}
}

// IsObject returns true if j is a JSON object (map).
func IsObject(j JSON) bool {
	_, ok := j.(map[string]JSON)
	return ok
}

// IsArray returns true if j is a JSON array or list.
func IsArray(j JSON) bool {
	_, ok := j.([]JSON)
	return ok
}

// IsStructure returns true if j is a JSON object or array.
func IsStructure(j JSON) bool {
	switch j.(type) {
	case map[string]JSON, []JSON:
		return true
	default:
		return false
	}
}

// eqArrayS returns true if array a as a string equals b.
// This represents Array.prototype.toString, reduced to the cases that can be true.
func eqArrayS(a []JSON, b string) bool {
	switch len(a) {
	case 0:
		return b == ""
	case 1:
		return cvs(a[0]) == b
	default:
		return false
	}
}

// eqArrayN returns true if array a as a string converted to a Number equals b.
// This represents Array.prototype.toString, reduced to the cases that can be true.
func eqArrayN(a []JSON, b JSON) bool {
	switch len(a) {
	case 0:
		return eqNum(b, 0.0)
	case 1:
		return eqNum(a[0], b)
	default:
		return false
	}
}

// badNum returns true if the string s would result in NaN in JS, where fp requires floating-point.
func badNum(s string, fp bool) bool {
	if s == "" {
		return false
	}
	if !fp {
		_, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return false
		}
	}
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return true
	}
	return false
}

// eqNum returns true if numeric values a and b are equal
func eqNum(a, b JSON) bool {
	if isFloat(a) || isFloat(b) {
		var zero float64
		va := cvf(a)
		vb := cvf(b)
		if va == 0.0 && vb == -zero || vb == 0.0 && va == -zero {
			return true
		}
		return va == vb // this will be false for NaN
	}
	return cvi(a) == cvi(b)
}

// eqVal returns the value of the Abstract Equality Comparison Algorithm (ECMA-262, 5.1, 11.9.3) [see notes/abstract-equality.pdf].
func eqVal(a, b JSON) bool {
	ta := typeOf(a)
	tb := typeOf(b)
	if ta != tb {
		// types differ, consider implicit conversions
		switch ta<<8 | tb {
		case Null<<8 | Undefined, Undefined<<8 | Null:
			// 11.9.3(2), 11.9.3(3)
			return true
		case Number<<8 | String:
			// 11.9.3(4)
			if badNum(cvs(b), isFloat(a)) {
				return false
			}
			return eqNum(a, b)
		case String<<8 | Number:
			// 11.9.3(5)
			if badNum(cvs(a), isFloat(b)) {
				return false
			}
			return eqNum(a, b)
		case Boolean<<8 | Number, Number<<8 | Boolean:
			// 11.9.3(6)
			return eqNum(a, b)
		case Boolean<<8 | String:
			// 11.9.3(6)
			if badNum(cvs(b), true) {
				return false
			}
			return cvf(a) == cvf(b) // wat!
		case String<<8 | Boolean:
			// 11.9.3(7)
			if badNum(cvs(a), true) {
				return false
			}
			return cvf(a) == cvf(b) // wat!
		case Object<<8 | Boolean:
			// 11.9.3(6) [a == ToNumber(b)]
			a, ok := a.([]JSON)
			b := cvi(b)
			return ok && (b == 0 && len(a) == 0 || len(a) == 1 && cvi(a[0]) == b)
		case Boolean<<8 | Object:
			// 11.9.3(7) [ToNumber(a) == b]
			b, ok := b.([]JSON)
			a := cvi(a)
			return ok && (a == 0 && len(b) == 0 || len(b) == 1 && cvi(b[0]) == a)
		case String<<8 | Object, Object<<8 | String:
			// 11.9.3(8), 11.9.3(9)
			// note only one can be an Object, and only when that's an array can == be true
			if a, ok := a.([]JSON); ok {
				return eqArrayS(a, cvs(b))
			}
			if b, ok := b.([]JSON); ok {
				return eqArrayS(b, cvs(a))
			}
			return false
		case Number<<8 | Object, Object<<8 | Number:
			// 11.9.3(8), 11.9.3(9)
			// note only one can be an Object, and only when that's an array can == be true
			if a, ok := a.([]JSON); ok {
				return eqArrayN(a, b)
			}
			if b, ok := b.([]JSON); ok {
				return eqArrayN(b, a)
			}
			return false
		default:
			return false
		}
	}
	// 11.9.3(1)
	switch ta {
	case Undefined, Null:
		return true
	case Number:
		return eqNum(a, b)
	case String:
		return a.(string) == b.(string)
	case Boolean:
		return a.(bool) == b.(bool)
	default:
		// unlike JavaScript, compare values not references
		switch a := a.(type) {
		case []JSON:
			b, ok := b.([]JSON)
			if !ok || len(a) != len(b) {
				return false
			}
			for i, ea := range a {
				if !eqVal(ea, b[i]) {
					return false
				}
			}
			return true
		case map[string]JSON:
			b, ok := b.(map[string]JSON)
			if !ok || len(a) != len(b) {
				return false
			}
			for k, v := range a {
				if !eqVal(v, b[k]) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
}

// searchJSON searches an array of values (treated as a list) for an instance of value v,
// returning f if found and !f otherwise.
// The appropriate equality function is used for the element type.
func searchJSON(vals []JSON, v JSON, f bool) bool {
	for _, el := range vals {
		if eqVal(el, v) {
			return f
		}
	}
	return !f
}

// convert a JavaScript primitive value, or array of primitive values, to a string
func cvs(v JSON) string {
	switch v := v.(type) {
	case error:
		return "undefined"
	case nil:
		return "null"
	case string:
		return v
	case bool, int, int64, float64:
		return fmt.Sprint(v)
	default:
		// non-primitive value, shouldn't be used
		return "[object Object]"
	}
}

// convert a value to integer.
func cvi(v JSON) int64 {
	// TO DO: protect against conversion traps
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v) // TO DO: traps if not representable?
	case string:
		if v == "" {
			return 0 // wat!
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// should really be NaN, so check first with badNum when required
			return 0
		}
		return n
	case IntVal: // appears in Slice (via OpBounds)
		return v.V()
	default:
		return 0
	}
}

// convert a value to floating-point.
func cvf(v JSON) float64 {
	// TO DO: protect against conversion traps
	switch v := v.(type) {
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		if v == "" {
			return 0.0 // wat!
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	default:
		//fmt.Printf("cvf(%#v)", v)
		return math.NaN()
	}
}

// convert a JSON value to boolean, the "truthy" JavaScript way: if it's not "falsy" it's true.
func cvb(v JSON) bool {
	switch v := v.(type) {
	case nil, error:
		return false
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		var zero float64
		return !math.IsNaN(v) && v != 0.0 && v != -zero
	case string:
		return v != ""
	default:
		//fmt.Printf("cvb: DEFAULT: %#v\n", v)
		return true
	}
}

// JavaScript types
type jsType int

const (
	Undefined jsType = iota
	Null
	Number
	String
	Boolean
	Object
)

// typeOf returns the  for a value, for use in Abstract Equality Comparison.
func typeOf(js JSON) jsType {
	switch js.(type) {
	case error:
		return Undefined
	case nil:
		return Null
	case int, int64, float64:
		return Number
	case string:
		return String
	case bool:
		return Boolean
	default:
		return Object
	}
}

// jsonString returns a flattened string representation of js (newlines replaced by spaces).
func jsonString(js JSON) string {
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	err := enc.Encode(js)
	if err != nil {
		return "!" + err.Error()
	}
	s := sb.String()
	l := len(s)
	if l > 0 && s[l-1] == '\n' {
		s = s[0 : l-1]
	}
	return strings.ReplaceAll(s, "\n", " ")
}
