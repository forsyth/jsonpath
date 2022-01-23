package JSONPath

import (
	"math"
	"testing"
)

type args []JSON
type eqTest struct {
	arg	args
	result	bool
}

var zero float64

// pairs lists operands to eqVal and the required result.
// Each pair is run for both orders of the arguments.
var pairs = []eqTest {
	// 11.9.3(1)

	// Undefined
	{args{nothing, nothing}, true},

	// Null
	{args{nil, nil}, true},

	// Number
	{args{42, 42.0}, true},
	{args{0, 0.0}, true},
	{args{0, 0}, true},
	{args{0.0, 0.0}, true},
	{args{0.0, -zero}, true},
	{args{math.NaN(), math.NaN()}, false},
	{args{1.5, math.NaN()}, false},

	// String
	{args{"", ""}, true},
	{args{"", "x"}, false},
	{args{"abc", "abc"}, true},
	{args{"abc", "abd"}, false},
	{args{"abc", "abcd"}, false},
	{args{"áé", "áé"}, true},
	{args{"ábc", "abc"}, false},

	// Boolean
	{args{true, true}, true},
	{args{true, false}, false},

	// 11.9.3(2)
	// Null, Undefined
	{args{nil, nothing}, true},

	// 11.9.3(3)
	// Undefined, Null
	{args{nothing, nil}, true},

	// 11.9.3(4), 11.9.3(5)
	// Number, String; String, Number
	{args{42, "42"}, true},
	{args{42, "42a"}, false},
	{args{0, ""}, true},	// wat!
	{args{1, ""}, false},
	{args{-42, "-42"}, true},
	{args{42.5, "42.5"}, true},
	{args{42.5, "42.5a"}, false},
	{args{0.0, ""}, true},	// wat!
	{args{1.0, ""}, false},
	{args{-42.5, "-42.5"}, true},
	{args{-zero, "-"}, false},
	{args{-zero, "-0.0"}, true},

	// 11.9.3(6), 11.9.3(7),
	// Boolean, T; T, Boolean
	// where the Boolean is converted ToNumber, hence surprises.

	// Boolean :: Number
	{args{false, 0}, true},
	{args{true, 1}, true},
	{args{false, 0.0}, true},
	{args{true, 1.0}, true},
	{args{true, 2}, false},

	// Boolean :: String
	{args{false, ""}, true},	// wat!
	{args{false, "0"}, true},	// wat!
	{args{false, "0.0"}, true},	// wat!
	{args{false, "false"}, false},	// wat!
	{args{false, "true"}, false},
	{args{true, ""}, false},
	{args{true, "1"}, true},	// wat!
	{args{true, "1.0"}, true},	// wat!
	{args{true, "2"}, false},
	{args{true, "true"}, false},	// wat!
	{args{true, "false"}, false},

	// Boolean :: Object
	{args{false, []JSON{}}, true},	// wat!
	{args{false, []JSON{false}}, true},	// wat!
	{args{true, []JSON{"false"}}, false},
	{args{false, make(map[string]JSON)}, false},
	{args{true, []JSON{}}, false},
	{args{true, []JSON{true}}, true},	// wat!
	{args{true, []JSON{"true"}}, false},
	{args{true, make(map[string]JSON)}, false},

	// 11.9.3(8), 11.9.3(9)
	// String :: Object, Number :: Object
	{args{"hello", []JSON{"hello"}}, true},	// wat!
	{args{"", []JSON{""}}, true},	// wat!
	{args{"", []JSON{}}, true},	// wat!
	{args{42, []JSON{"42"}}, true},	// wat!
	{args{42, []JSON{42}}, true},	// wat!
	{args{42, map[string]JSON{"42": true}}, false},
	{args{42.0, []JSON{"42.0"}}, true},	// wat!
	{args{42.0, []JSON{42.0}}, true},	// wat!
	{args{42.0, map[string]JSON{"42.0": true}}, false},
}

// TestEquality runs through a set of tests of the abstract equality comparison algorithm (JS ==).
func TestEquality(t *testing.T) {
	for i, p := range pairs {
		for j := 0; j < 2; j++ {
			a := p.arg[j]
			b := p.arg[^j & 1]
			r := eqVal(a, b)
			if r != p.result {
				t.Errorf("pair %d.%d: %#v == %#v: want %#v got %#v", i, j, a, b, p.result, r)
			}
		}
	}
}
