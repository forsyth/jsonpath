package JSONPath

import (
	"encoding/json"
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

// eqVal returns true if value(a) == value(b) in the expression language.
func eqVal(a, b JSON) bool {
	// this seems to be easier to follow than nested type switches
	if isString(b) && isString(a) {
		return cvs(b) == cvs(a)
	}
	if isFloat(b) && isFloat(a) || isFloat(b) && isInt(a) || isInt(b) && isFloat(a) {
		return cvf(b) == cvf(a)
	}
	if isInt(b) && isInt(a) {
		return cvi(b) == cvi(a)
	}
	// let this one be truthy on the LHS
	switch b := b.(type) {
	case bool:
		return b == cvb(a)
	default:
		return false
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

// return the value of a string, or the empty string
func cvs(v JSON) string {
	switch v := v.(type) {
	case string:
		return v
	default:
		// TO DO
		return ""
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
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
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
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0.0
		}
		return f
	default:
		//fmt.Printf("cvf(%#v)", v)
		return 0.0
	}
}

// convert a JSON value to boolean, the "truthy" JavaScript way.
func cvb(v JSON) bool {
	switch v := v.(type) {
	case nil, error:
		return false
	case bool:
		return v
	case string:
		return v != ""
	case []JSON:
		return len(v) != 0
	case map[string]JSON:
		return len(v) != 0
	default:
		//fmt.Printf("cvb: DEFAULT: %#v\n", v)
		return true
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
