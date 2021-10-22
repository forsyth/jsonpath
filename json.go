package JSONPath

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
