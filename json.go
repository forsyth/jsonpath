package JSONPath

// JSON is a synonym for the interface{} structures returned by encoding/json,
// used as values in the JSON machine, to make it clear that's what they are.
type JSON = interface{}

// enquiry functions on JSON, sometimes easier to read than type switches

func isString(v JSON) bool {
	_, ok := v.(string)
	return ok
}

func isInt(v JSON) bool {
	switch v.(type) {
	case int, int64:
		return true
	default:
		return false
	}
}

func isFloat(v JSON) bool {
	_, ok := v.(float64)
	return ok
}

func isSlice(v JSON) bool {
	_, ok := v.(*Slice)
	return ok
}

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

func isSimple(v JSON) bool {
	switch v.(type) {
	case bool, int, int64, float64, string:
		return true
	default:
		return false
	}
}
