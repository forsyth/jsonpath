package mach

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
	//	"github.com/forsyth/jsonpath/paths"
)

// Function represents a predefined function with na args (or AnyNumber) with body fn.
type Function struct {
	na int
	fn func([]JSON) JSON
}

// AnyNumber as Function.na means any number of args.
const AnyNumber = -1

// functions is the default set of predefined functions.
var functions = map[string]Function{
	"abs": {
		1,
		func(args []JSON) JSON {
			switch n := args[0].(type) {
			case int:
				w := int64(n)
				if w < 0 {
					return -w
				}
				return w
			case int64:
				if n < 0 {
					n = -n
					if n < 0 {
						return ErrOverflow
					}
				}
				return n
			case float64:
				if n < 0 {
					return -n
				}
				return n
			default:
				return ErrType
			}
		},
	},
	"avg": {
		1,
		func(args []JSON) JSON {
			if a, ok := args[0].([]JSON); ok {
				var sum float64
				for _, v := range a {
					if !isArith(v) {
						return ErrType
					}
					sum += cvf(v)
				}
				return sum / float64(len(a))
			}
			return ErrType
		},
	},
	"ceil": {
		1,
		func(args []JSON) JSON {
			switch f := args[0].(type) {
			case int, int64:
				return f
			case float64:
				return math.Ceil(f)
			default:
				return ErrType
			}
		},
	},
	"contains": {
		2,
		func(args []JSON) JSON {
			switch a := args[0].(type) {
			case string:
				// does string a contain args[1]?
				if b, ok := args[1].(string); ok {
					return strings.Contains(a, b)
				}
				return ErrType
			case []JSON:
				// look for value args[1] in slice a
				for _, v := range a {
					if eqVal(v, args[1]) {
						return true
					}
				}
				return false
			default:
				return ErrType
			}
		},
	},
	"ends_with": {
		2,
		func(args []JSON) JSON {
			a, b, ok := stringArgs(args)
			if !ok {
				return nothing
			}
			return strings.HasSuffix(a, b)
		},
	},
	"floor": {
		1,
		func(args []JSON) JSON {
			switch f := args[0].(type) {
			case int, int64:
				return f
			case float64:
				return math.Floor(f)
			default:
				return ErrType
			}
		},
	},
	"keys": {
		1,
		func(args []JSON) JSON {
			if obj, ok := args[0].(map[string]JSON); ok {
				keys := make([]JSON, 0, len(obj))
				for k, _ := range obj {
					keys = append(keys, k)
				}
				return keys
			} else {
				return ErrType
			}
		},
	},
	"length": {
		1,
		func(args []JSON) JSON {
			switch a := args[0].(type) {
			case string:
				return int64(utf8.RuneCountInString(a))
			case []JSON:
				return int64(len(a))
			case map[string]JSON:
				return int64(len(a))
			default:
				return nil
			}
		},
	},
	"max": {
		AnyNumber,
		func(args []JSON) JSON {
			if len(args) == 0 {
				return nil
			}
			if len(args) == 1 {
				// min(array)
				if a, ok := args[0].([]JSON); ok {
					return minMaxArray(a, maxArith, maxString)
				}
				return ErrType
			}
			// min(a, b, ...)
			return minMaxArray(args, maxArith, maxString)
		},
	},
	"min": {
		AnyNumber,
		func(args []JSON) JSON {
			if len(args) == 0 {
				return nil
			}
			if len(args) == 1 {
				// min(array)
				if a, ok := args[0].([]JSON); ok {
					return minMaxArray(a, minArith, minString)
				}
				return ErrType
			}
			// min(a, b, ...)
			return minMaxArray(args, minArith, minString)
		},
	},
	"prod": {
		1,
		func(args []JSON) JSON {
			if a, ok := args[0].([]JSON); ok {
				if len(a) == 0 {
					return nil
				}
				return arithArrayOp(a, func(index int, result, el float64) float64 {
					if index > 0 {
						return result * el
					}
					return el
				})
			}
			return ErrType
		},
	},
	"starts_with": {
		2,
		func(args []JSON) JSON {
			a, b, ok := stringArgs(args)
			if !ok {
				return nothing
			}
			return strings.HasPrefix(a, b)
		},
	},
	"sum": {
		1,
		func(args []JSON) JSON {
			if a, ok := args[0].([]JSON); ok {
				if len(a) == 0 {
					return float64(0.0)
				}
				return arithArrayOp(a, func(index int, result, el float64) float64 {
					return result + el
				})
			}
			return ErrType
		},
	},
	"to_number": {
		1,
		func(args []JSON) JSON {
			switch a := args[0].(type) {
			case int, int64, float64:
				return a
			case string:
				v, err := strconv.ParseFloat(a, 64)
				if err != nil {
					return err
				}
				return v
			default:
				return ErrType
			}
		},
	},
	"tokenize": {
		2,
		func(args []JSON) JSON {
			s, re, ok := stringArgs(args)
			if !ok {
				return ErrType
			}
			prog, err := regexp.Compile(re)
			if err != nil {
				return err
			}
			matches := prog.Split(s, -1)
			result := []JSON{}
			for _, m := range matches {
				result = append(result, m)
			}
			return result
		},
	},
}

func stringArgs(args []JSON) (string, string, bool) {
	if len(args) != 2 {
		return "", "", false
	}
	var a, b string
	switch s := args[0].(type) {
	case string:
		a = s
	default:
		return "", "", false
	}
	switch s := args[1].(type) {
	case string:
		b = s
	default:
		return "", "", false
	}
	return a, b, true
}

func arithArrayOp(a []JSON, f func(int, float64, float64) float64) JSON {
	if len(a) == 0 {
		return nil
	}
	var result float64
	for i, v := range a {
		if !isArith(v) {
			return ErrType
		}
		result = f(i, result, cvf(v))
	}
	return result
}

func stringArrayOp(a []JSON, f func(int, string, string) string) JSON {
	if len(a) == 0 {
		return nil
	}
	var result string
	for i, v := range a {
		if s, ok := v.(string); ok {
			result = f(i, result, s)
		} else {
			return ErrType
		}
	}
	return result
}

func minMaxArray(a []JSON, arithf func(int, float64, float64) float64, stringf func(int, string, string) string) JSON {
	if len(a) == 0 {
		return nil
	}
	switch a[0].(type) {
	case int, int64, float64:
		return arithArrayOp(a, arithf)
	case string:
		return stringArrayOp(a, stringf)
	default:
		return ErrType
	}
}

func minArith(index int, min, el float64) float64 {
	if index == 0 || el < min {
		return el
	}
	return min
}

func maxArith(index int, max, el float64) float64 {
	if index == 0 || el > max {
		return el
	}
	return max
}

func minString(index int, min, el string) string {
	if index == 0 || el < min {
		return el
	}
	return min
}

func maxString(index int, max, el string) string {
	if index == 0 || el < max {
		return el
	}
	return max
}
