package JSONPath

// "When I say 'run!', run!"

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"unicode/utf8"
)

var (
	ErrNoKey     = errors.New("no such key")
	ErrBadIndex  = errors.New("invalid array index")
	ErrNotArray  = errors.New("not an array")
	ErrNotObject = errors.New("not an object")
	ErrFailure   = errors.New("failed")
	ErrType      = errors.New("operand or parameter has wrong type")
	ErrOverflow  = errors.New("arithmetic overflow")
)

// machine is the current state of the virtual machine.
type machine struct {
	prog    *Program
	root    JSON          // $
	out     []JSON        // current set of output values
	dot     JSON          // @ in a filter
	stack   []JSON        // expression stack
	sp      int           // expression stack pointer
	pc      int           // next instruction
	values  []<-chan JSON // values from OpFor for OpFilter or OpNest
	tracing bool
}

func (m *machine) push(val JSON) {
	if m.sp >= len(m.stack) {
		m.stack = append(m.stack, val)
		m.sp = len(m.stack)
	} else {
		m.stack[m.sp] = val
		m.sp++
	}
}

func (m *machine) pop() JSON {
	if m.sp == 0 {
		panic("stack underflow")
	}
	m.sp--
	v := m.stack[m.sp]
	m.stack[m.sp] = nil
	return v
}

func (m *machine) popN(n int64) []JSON {
	if int64(m.sp) < n {
		panic("stack underflow")
	}
	m.sp -= int(n)
	a := make([]JSON, n)
	copy(a, m.stack)
	for i := 0; i < int(n); i++ {
		m.stack[i] = nil
	}
	return a
}

func (m *machine) branch(pc int) {
	if pc <= 0 || pc > len(m.prog.orders) {
		panic(fmt.Sprintf("branch pc out of range: %d at pc %d", pc, m.pc-1))
	}
	//fmt.Printf("-- branch %d -> %d\n", m.pc-1, pc)
	m.pc = pc
}

func (m *machine) pushInput(c <-chan JSON) {
	m.values = append(m.values, c)
}

func (m *machine) topInput() <-chan JSON {
	return m.values[len(m.values)-1]
}

func (m *machine) popInput() {
	m.values = m.values[0 : len(m.values)-1]
}

// nothing represents an evaluation without a usable result.
// Any error will do.
// ("nothing is better than Advil")
var nothing = ErrFailure

// isNothing returns true if v represents an evaluation failure.
func isNothing(v JSON) bool {
	_, ok := v.(error)
	return ok
}

// valOK checks that v is something and returns true if so.
// Otherwise it pushes nothing and returns false.
func (m *machine) valOK(v JSON) bool {
	if isNothing(v) {
		m.push(nothing)
		return false
	}
	return true
}

// valsOK checks that neither a nor b is nothing and returns true if so.
// Otherwise it pushes nothing and returns false.
func (m *machine) valsOK(a, b JSON) bool {
	if isNothing(a) || isNothing(b) {
		m.push(nothing)
		return false
	}
	return true
}

// Run applies the current Program to the root of a JSON structure, returning a collection of JSON structures from it (which might be empty) as selected by the original path expression, or a fatal run-time error.
// Following the usual JavaScript conventions, many errors are not fatal, but yield a null result.
func (p *Program) Run(root JSON) ([]JSON, error) {
	vm := &machine{prog: p, root: root, out: []JSON{root}, pc: 0, tracing: false}
	for vm.pc < len(p.orders) {
		ord := p.orders[vm.pc]
		vm.pc++
		switch ord.op() {
		// leaf operations
		case OpInt:
			if ord.isSmallInt() {
				vm.push(ord.smallInt())
				break
			}
			vm.push(p.value(ord.index()).(Valuer).Value())
		case OpBool:
			if !ord.isSmallInt() {
				panic("require smallInt")
			}
			v := ord.smallInt()
			if v != 0 {
				vm.push(true)
			} else {
				vm.push(false)
			}
		case OpNull:
			vm.push(nil)
		case OpReal, OpRE, OpBounds, OpString:
			vm.push(p.value(ord.index()).(Valuer).Value())
		case OpID:
			vm.push(p.value(ord.index()))
		case OpExp:
			// expression in path is either string or integer (a key or index);
			// other values are converted to integer.
			v := vm.pop()
			if !vm.valOK(v) {
				break
			}
			if isString(v) || isInt(v) {
				vm.push(v)
			} else {
				vm.push(cvi(v))
			}

		// path operations, working on each member of the current output set
		case OpWild:
			vm.out = applySelection(vm.out, func(val JSON, acc []JSON) []JSON {
				return valsWild(acc, val)
			})
		case OpMember, OpSelect:
			negIndex := ord.op() == OpSelect // only [] can index from end of array
			sel := vm.pop()                  // can be ID, String, Int, Expr(result) or Slice
			if isNothing(sel) {
				vm.out = []JSON{}
				break
			}
			vm.out = applySelection(vm.out, func(val JSON, acc []JSON) []JSON {
				return valsByKey(acc, val, sel, negIndex)
			})
		case OpUnion:
			// note that it's (apparently) a union that yields a bag, not a set
			n := ord.smallInt()
			sels := vm.popN(n)
			vm.out = applySelection(vm.out, func(val JSON, acc []JSON) []JSON {
				for _, sel := range sels {
					if !isNothing(sel) {
						acc = valsByKey(acc, val, sel, true)
					}
				}
				return acc
			})

		// path operations, working on the value in dot
		case OpFilter, OpNestFilter:
			v := vm.pop()
			//fmt.Printf("FILTER: %#v\n", v)
			if !isNothing(v) && cvb(v) {
				vm.out = append(vm.out, vm.dot)
			}
		case OpNestWild:
			vm.out = valsWild(vm.out, vm.dot)
		case OpNestMember, OpNestSelect:
			negIndex := ord.op() == OpNestSelect // only [] can index from end of array
			sel := vm.pop()                      // can be ID, String, Int, Expr(result) or Slice
			if !isNothing(sel) {
				vm.out = valsByKey(vm.out, vm.dot, sel, negIndex)
			}
		case OpNestUnion:
			// note that it's (apparently) a union that yields a bag, not a set
			n := ord.smallInt()
			sels := vm.popN(n)
			for _, sel := range sels {
				if !isNothing(sel) {
					vm.out = valsByKey(vm.out, vm.dot, sel, true)
				}
			}

		// iterating over members of current vm.out directly (OpFor) and all their descendents (OpNest)
		case OpFor:
			looptop(vm, stepping, ord.pc())
		case OpNest:
			looptop(vm, walker, ord.pc())
		case OpRep:
			js, more := <-vm.topInput()
			if !more {
				//fmt.Printf("rep: all done\n")
				vm.popInput()
				vm.dot = nil
				break
			}
			vm.dot = js
			//fmt.Printf("rep: next: %v\n", vm.dot)
			vm.branch(ord.pc())

		// expression operators
		case OpRoot:
			vm.push(vm.root)
		case OpCurrent:
			vm.push(vm.dot)
		case OpDot:
			sel := vm.pop()
			val := vm.pop()
			if !vm.valsOK(sel, val) {
				break
			}
			key := sel.(NameVal)
			if !vm.valOK(key) {
				break
			}
			if key.S() == "length" {
				n := -1
				switch val := val.(type) {
				case []JSON:
					n = len(val)
				case map[string]JSON:
					n = len(val)
				case string:
					n = utf8.RuneCountInString(val)
				}
				if n < 0 {
					//fmt.Printf(".length of non-array/object: %#v\n", val)
					vm.push(nothing)
					break
				}
				vm.push(int64(n))
				break
			}
			//fmt.Printf("Dot: %#v . %s", val, key.S())
			if el, ok := val.(map[string]JSON); ok {
				fv, ok := valByKey(el, sel, false)
				//fmt.Printf(" -> %#v\n", fv)
				if !ok {
					vm.push(nothing)
					break
				}
				vm.push(fv)
			} else {
				//fmt.Printf(" -> . of non-object: %#v %#v\n", el, key)
				vm.push(nothing)
			}
		case OpIndex:
			// array[index] or obj['field']
			index := vm.pop()
			val := vm.pop()
			if !vm.valsOK(val, index) {
				break
			}
			switch val := val.(type) {
			case []JSON:
				sel := cvi(index) // can be only int, or convertible
				//fmt.Printf("index=%#v\n", sel)
				if sel < 0 || sel >= int64(len(val)) {
					vm.push(nothing)
					break
				}
				vm.push(val[sel])
			case map[string]JSON:
				// handles both JSON objects and arrays
				res, ok := valByKey(val, index, true)
				if !ok {
					vm.push(nothing)
					break
				}
				vm.push(res)
			default:
				//fmt.Printf("index of %#v by %#v\n", val, index)
				vm.push(nothing)
			}
		case OpSlice:
			b := vm.pop()
			a := vm.pop()
			if isNothing(a) || isNothing(b) {
				vm.push(nothing)
				break
			}
			slice, ok1 := b.(*Slice)
			array, ok2 := a.([]JSON)
			if !ok1 || !ok2 {
				//fmt.Printf("slice failed: %#v %#v\n", array, slice)
				vm.push([]JSON{})
				break
			}
			//fmt.Printf("slice=%v %#v\n", slice, array)
			vm.push(slicing(array, slice))
		case OpOr:
			b := vm.pop()
			a := vm.pop()
			if isNothing(a) && isNothing(b) {
				vm.push(nothing)
				break
			}
			if !isNothing(a) && cvb(a) {
				vm.push(a)
				break
			}
			vm.push(b)
		case OpAnd:
			b := vm.pop()
			a := vm.pop()
			if isNothing(a) || !cvb(a) {
				vm.push(a)
				break
			}
			vm.push(b)
		case OpAdd:
			// TO DO: allow string+string concatenation?
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			vm.push(arith(a, b, func(i, j int64) int64 { return i + j }, func(x, y float64) float64 { return x + y }))
		case OpSub:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			vm.push(arith(a, b, func(i, j int64) int64 { return i - j }, func(x, y float64) float64 { return x - y }))
		case OpMul:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			vm.push(arith(a, b, func(i, j int64) int64 { return i * j }, func(x, y float64) float64 { return x * y }))
		case OpDiv:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			vm.push(divide(a, b, func(i, j int64) int64 { return i / j }, func(x, y float64) float64 { return x / y }))
		case OpMod:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			vm.push(divide(a, b, func(i, j int64) int64 { return i / j }, func(x, y float64) float64 { return math.Mod(x, y) }))
		case OpNeg:
			v := vm.pop()
			if !vm.valOK(v) {
				break
			}
			switch v.(type) {
			case float64:
				vm.push(-v.(float64))
			default:
				vm.push(-cvi(v))
			}
		case OpNot:
			vm.push(!cvb(vm.pop()))
		case OpEQ:
			b := vm.pop()
			a := vm.pop()
			vm.push(eqVal(a, b))
		case OpNE:
			b := vm.pop()
			a := vm.pop()
			vm.push(!eqVal(a, b))
		case OpLT:
			b := vm.pop()
			a := vm.pop()
			vm.push(relation(a, b, func(i, j int64) bool { return i < j },
				func(x, y float64) bool { return x < y }, func(s, t string) bool { return s < t }))
		case OpLE:
			b := vm.pop()
			a := vm.pop()
			vm.push(relation(a, b, func(i, j int64) bool { return i <= j },
				func(x, y float64) bool { return x <= y }, func(s, t string) bool { return s <= t }))
		case OpGE:
			b := vm.pop()
			a := vm.pop()
			vm.push(relation(a, b, func(i, j int64) bool { return i >= j },
				func(x, y float64) bool { return x >= y }, func(s, t string) bool { return s >= t }))
		case OpGT:
			b := vm.pop()
			a := vm.pop()
			vm.push(relation(a, b, func(i, j int64) bool { return i > j },
				func(x, y float64) bool { return x > y }, func(s, t string) bool { return s > t }))
		case OpArray:
			n := ord.smallInt()
			vm.push(vm.popN(n))
		case OpMatch:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			var err error
			var re *regexp.Regexp
			switch b := b.(type) {
			case *regexp.Regexp:
				// already compiled
				re = b
			case string:
				// dynamic string value, to be compiled now
				re, err = regexp.CompilePOSIX(b)
				if err != nil {
					return nil, err // user visible so don't include pc
				}
			default:
				return nil, fmt.Errorf("%s requires string or /re/ right operand, not %#v", ord.op(), b)
			}
			switch s := a.(type) {
			case string:
				vm.push(re.MatchString(s))
			default:
				return nil, fmt.Errorf("%s requires string left operand, not %s", ord.op(), a)
			}
		case OpIn, OpNin:
			b := vm.pop()
			a := vm.pop()
			if !vm.valsOK(a, b) {
				break
			}
			switch b := b.(type) {
			case []JSON:
				vm.push(searchJSON(b, a, ord.op() == OpIn))
			default:
				return nil, fmt.Errorf("%s requires array right operand, not %s", ord.op(), b)
			}
		//case OpCall:
		default:
			return nil, fmt.Errorf("unimplemented %#v at pc %d", ord.op(), vm.pc-1)
		}
		if vm.tracing {
			fmt.Printf("%#v ->\n", ord.op())
			fmt.Printf("\t[")
			for i, x := range vm.out {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", jsonString(x))
			}
			fmt.Printf("]\n")
		}
	}
	if vm.out == nil {
		return []JSON{}, nil
	}
	return vm.out, nil
}

// apply runs the selection function f on each element of the src array, returning a new array with the results.
func applySelection(src []JSON, f func(JSON, []JSON) []JSON) []JSON {
	vals := []JSON{}
	for _, el := range src {
		vals = f(el, vals)
	}
	return vals
}

// looptop sets up iteration (OpFor, OpNest) over a set of values produced by the producer process.
func looptop(vm *machine, producer func(chan<- JSON, []JSON), epc int) {
	if len(vm.out) == 0 {
		//fmt.Printf("loop: empty out\n")
		vm.branch(epc)
		return
	}
	// TO DO: special case len(vm.out) == 1, just set vm.dot
	values := make(chan JSON)
	vm.pushInput(values)
	go producer(values, vm.out)
	vm.out = []JSON{}
	vm.dot = <-values
}

// sliceEval returns an interpretation of the given Slice with respect to an array of length l.
// A negative index is converted into an offset from the array's end.
// A negative stride means the selected entries are to be in reverse order (start down to end, exclusive).
func sliceEval(slice *Slice, l int64) (int64, int64, int64) {
	stride := int64(1)
	if slice.Stride != nil {
		stride = cvi(slice.Stride)
	}
	start := int64(0)
	end := int64(l)
	if stride < 0 {
		start = l - 1
		end = -l - 1 // ie stop just beyond start of array, descending
	}
	if slice.Start != nil {
		start = cvi(slice.Start)
	}
	if slice.End != nil {
		end = cvi(slice.End)
	}
	if stride < 0 {
		// items from start descending to end (excluded)
		switch {
		case start >= l:
			start = l - 1
		case start < 0:
			start += l
			if start < 0 {
				// no items below start
				start = -1
			}
		}
		switch {
		case end > l:
			end = l
		case end < 0:
			end += l
			if end < 0 {
				end = -1
			}
		}
		//fmt.Printf("%d @ (%d, %d, %d)\n", l, start, end, stride)
		return start, end, stride
	}
	// items [start, end) in the usual way
	switch {
	case start > l:
		start = l
	case start < 0:
		// distance from end of the array (-1 is last item)
		start += l
		if start < 0 {
			start = 0
		}
	}
	switch {
	case end > l:
		end = l
	case end < 0:
		end += l
		if end < 0 {
			end = 0
		}
	}
	return start, end, stride
}

// arith decides whether to do an arithmetic operation as int or float, and returns the resulting value.
// TO DO: could just do all expression arithmetic in float64?
func arith(a, b JSON, intf func(int64, int64) int64, floatf func(float64, float64) float64) JSON {
	if isNothing(a) || isNothing(b) {
		return nothing
	}
	if !(isArith(a) && isArith(b)) {
		return nothing
	}
	if isFloat(a) || isFloat(b) {
		return floatf(cvf(a), cvf(b))
	}
	return intf(cvi(a), cvi(b))
}

// divide decides whether to do a division operation as int or float, and returns the resulting value.
// Division by zero yields nothing (should probably be true for NaN as well).
func divide(a, b JSON, intf func(int64, int64) int64, floatf func(float64, float64) float64) JSON {
	if isNothing(a) || isNothing(b) {
		return nothing
	}
	if !(isArith(a) && isArith(b)) {
		return nothing
	}
	if isFloat(a) || isFloat(b) {
		bf := cvf(b)
		if bf == 0.0 {
			return nothing
		}
		return floatf(cvf(a), cvf(b))
	}
	bi := cvi(b)
	if bi == 0 {
		return nothing
	}
	return intf(cvi(a), bi)
}

// relation decides whether to do a comparison operation as int or float, and returns the resulting value.
func relation(a, b JSON, intf func(int64, int64) bool, floatf func(float64, float64) bool, stringf func(string, string) bool) JSON {
	if isNothing(a) || isNothing(b) {
		return nothing
	}
	if !(isSimple(a) && isSimple(b)) {
		return false
	}
	if isString(a) || isString(b) {
		if !(isString(a) && isString(b)) {
			return false
		}
		return stringf(cvs(a), cvs(b))
	}
	if isFloat(a) || isFloat(b) {
		//fmt.Printf("REL2 %#v %#v XX %v %v %v\n", cvf(a), cvf(b), isFloat(a), isFloat(b), isInt(b))
		return floatf(cvf(a), cvf(b))
	}
	//fmt.Printf("REL1 %#v %#v\n", cvi(a), cvi(b))
	return intf(cvi(a), cvi(b))
}

// slicing returns the slice of src.
func slicing(src []JSON, slice *Slice) []JSON {
	start, end, stride := sliceEval(slice, int64(len(src)))
	switch {
	case stride == 1:
		return src[start:end]
	case stride > 0:
		vals := []JSON{}
		for i := start; i < end; i += stride {
			vals = append(vals, src[i])
		}
		return vals
	case stride < 0:
		vals := []JSON{}
		for i := start; i > end; i += stride {
			vals = append(vals, src[i])
		}
		return vals
	default: // stride == 0, inoperative
		return []JSON{}
	}
}

// valsWild adds to vals the members of objects and elements of arrays in src.
func valsWild(vals []JSON, src JSON) []JSON {
	switch src := src.(type) {
	case []JSON:
		for _, el := range src {
			//fmt.Printf("el: %#v\n", el)
			vals = append(vals, el)
		}
	case map[string]JSON:
		for _, v := range src {
			vals = append(vals, v)
		}
	}
	return vals
}

// valsByKey adds to vals a set of values from the src that satisfy the given key (eg, member name, index, slice).
// TO DO: use a map to check whether the values have been seen when forming a union.
func valsByKey(vals []JSON, src JSON, key JSON, negIndex bool) []JSON {
	if isSlice(key) {
		src, ok := src.([]JSON)
		if !ok {
			return nil
		}
		slice := key.(*Slice)
		start, end, stride := sliceEval(slice, int64(len(src)))
		switch {
		case stride > 0:
			for i := start; i < end; i += stride {
				vals = append(vals, src[i])
			}
		case stride < 0:
			for i := start; i > end; i += stride {
				vals = append(vals, src[i])
			}
		case stride == 0:
			// could yield an error, but in the spirit of jsonPath, we'll do nothing
		}
		return vals
	}
	switch src := src.(type) {
	case []JSON:
		if isInt(key) {
			// [integer]
			v, ok := valByKey(src, key, negIndex)
			if ok {
				vals = append(vals, v)
			}
		}
		return vals
	case map[string]JSON:
		v, ok := valByKey(src, key, false)
		if ok {
			vals = append(vals, v)
		}
	default:
		// neither object nor array
	}
	return vals
}

// keyVal converts a key into a suitable string value to index a Go JSON map.
func mapKey(key JSON) string {
	switch key := key.(type) {
	case string:
		return key
	case float64:
		// floats here are the result of (expr),
		// and an integer is required.
		return fmt.Sprint(int64(float64(key)))
	case NameVal:
		return key.S()
	default:
		return fmt.Sprint(key)
	}
}

func valByKey(src JSON, key JSON, negIndex bool) (JSON, bool) {
	//fmt.Printf("%#v %s\n", src, key)
	switch src := src.(type) {
	case map[string]JSON:
		s := mapKey(key)
		v, ok := src[s]
		if !ok {
			//fmt.Printf(" -> nil\n")
			return nil, false
		}
		//fmt.Printf(" -> %#v\n", v)
		return v, true
	case []JSON:
		l := int64(len(src))
		n := cvi(key)
		if negIndex && n < 0 {
			n += l
		}
		if n >= 0 && n < l {
			//fmt.Printf(" -> %#v\n", src[n])
			return src[n], true
		}
		//fmt.Printf(" -> nil\n")
		return nil, false
	case nil:
		// null.key is null
		return src, true
	default:
		//fmt.Printf(" -> nil\n")
		return nil, false
	}
}

// stepping sends the JSON structures from the given array one at a time on values.
func stepping(values chan<- JSON, vals []JSON) {
	defer close(values)
	for _, v := range vals {
		switch v := v.(type) {
		case []JSON:
			for _, el := range v {
				values <- el
			}
		case map[string]JSON:
			for _, el := range v {
				values <- el
			}
		}
	}
}

// walker walks down a sequence of JSON structures passing object and array substructure back in values.
// The order is defined in 9.1.1.8 [[Descendants]] of
// https://www.ecma-international.org/wp-content/uploads/ECMA-357_2nd_edition_december_2005.pdf
func walker(values chan<- JSON, vals []JSON) {
	defer close(values)
	for _, v := range vals {
		if IsStructure(v) {
			walkdown(values, v)
		}
	}
}

func walkdown(values chan<- JSON, val JSON) {
	values <- val
	switch val := val.(type) {
	case map[string]JSON:
		// note object members, and walk down from each member that's an array or object
		for _, v := range val {
			if IsStructure(v) {
				walkdown(values, v)
			}
		}
	case []JSON:
		// elements
		for _, v := range val {
			if IsStructure(v) {
				walkdown(values, v)
			}
		}
	default:
		// bool, float64, string or nil
	}
}
