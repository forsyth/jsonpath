package JSONPath

// Op represents path expression leaf and expression operators, and filter and expression engine operators.
// Op values also are the tokens produced by the lexical analyser.
//
// Nest is tricky because it's a portmanteau, thanks to ".." "[" subscript "]".
// It can be applied to the full syntax of "subscript", including union.
// As a plain Op, uniquely it would be a Step Op with another Step as its Arg value,
// and having an OpNestFlag turned out to be clumsy too,
// whereas having OpNest variants is slightly tedious but avoids special cases.

type Op int

const (
	OpError  Op = iota // illegal token, deliberately the same as the zero value
	OpEof              // end of file
	OpID               // identifier
	OpString           // single- or double-quoted string
	OpInt              // integer
	OpReal             // real number (might be used in expressions)
	OpRE               // /re/

	// path operators
	OpRoot    // $
	OpCurrent // @
	OpDot     // .
	OpSelect  // [] when used for selection
	OpIndex   // [] when used for indexing
	OpSlice   // [lb: ub: stride] slice operator, Arg of OpIndex, OpUnion
	OpUnion   // [key1, key2 ...]
	OpWild    // *
	OpFilter  // ?(...)
	OpExp     // (...)

	// path nest operators
	OpNest       // .. member
	OpNestSelect // OpNest + OpSelect
	OpNestIndex  // OpNest + OpIndex
	OpNestUnion  // .. [key1, key2, ...]
	OpNestWild   // .. [*]
	OpNestFilter // .. [?(expr)]

	// expression operators, in both filters and "expression engines"
	OpLT    // <
	OpLE    // <=
	OpEQ    // = or ==
	OpNE    // !=
	OpGE    // >=
	OpGT    // >
	OpAnd   // &&
	OpOr    // ||
	OpMul   // *
	OpDiv   // /
	OpMod   // %
	OpNeg   // unary -
	OpAdd   // +
	OpSub   // binary -
	OpCall  // function call id(args)
	OpArray // [e-list]
	OpIn    // "in"
	OpNin   // "nin", not in
	OpMatch // ~= (why not just ~)
	OpNot   // unary !

	// operators internal to the polish notation and stack VM
	OpVal // operand
)

var opNames map[Op]string = map[Op]string{
	OpError:      "OpError",
	OpEof:        "OpEof",
	OpID:         "OpID",
	OpString:     "OpString",
	OpInt:        "OpInt",
	OpReal:       "OpReal",
	OpRE:         "OpRE",
	OpRoot:       "OpRoot",
	OpCurrent:    "OpCurrent",
	OpDot:        "OpDot",
	OpSelect:     "OpSelect",
	OpIndex:      "OpIndex",
	OpSlice:      "OpSlice",
	OpUnion:      "OpUnion",
	OpWild:       "OpWild",
	OpFilter:     "OpFilter",
	OpExp:        "OpExp",
	OpNest:       "OpNest",
	OpNestSelect: "OpNestSelect",
	OpNestIndex:  "OpNestIndex",
	OpNestUnion:  "OpNestUnion",
	OpNestWild:   "OpNestWild",
	OpNestFilter: "OpNestFilter",
	OpLT:         "OpLT",
	OpLE:         "OpLE",
	OpEQ:         "OpEQ",
	OpNE:         "OpNE",
	OpGE:         "OpGE",
	OpGT:         "OpGT",
	OpAnd:        "OpAnd",
	OpOr:         "OpOr",
	OpMul:        "OpMul",
	OpDiv:        "OpDiv",
	OpMod:        "OpMod",
	OpNeg:        "OpNeg",
	OpAdd:        "OpAdd",
	OpSub:        "OpSub",
	OpCall:       "OpCall",
	OpArray:      "OpArray",
	OpIn:         "OpIn",
	OpNin:        "OpNin",
	OpMatch:      "OpMatch",
	OpNot:        "OpNot",
	OpVal:        "OpVal",
}

var opText map[Op]string = map[Op]string{
	OpError:      "(error)",
	OpEof:        "(eof)",
	OpID:         "identifier",
	OpString:     "string",
	OpInt:        "integer",
	OpReal:       "real number",
	OpRE:         "regular expression",
	OpRoot:       "$",
	OpCurrent:    "@",
	OpDot:        ".",
	OpSelect:     "[]selection",
	OpIndex:      "[]index",
	OpSlice:      "[]slice",
	OpUnion:      "[]union",
	OpWild:       "*",
	OpFilter:     "?(filter)",
	OpExp:        "(exp)",
	OpNest:       "..",
	OpNestSelect: "..[[selection",
	OpNestIndex:  "..[]index",
	OpNestUnion:  "..[]union",
	OpNestWild:   "..*",
	OpNestFilter: "..$(filter)",
	OpLT:         "<",
	OpLE:         "<=",
	OpEQ:         "==",
	OpNE:         "!=",
	OpGE:         ">=",
	OpGT:         ">",
	OpAnd:        "&&",
	OpOr:         "||",
	OpMul:        "*",
	OpDiv:        "/",
	OpMod:        "%",
	OpNeg:        "unary -",
	OpAdd:        "+",
	OpSub:        "-",
	OpCall:       "function call",
	OpArray:      "array value",
	OpIn:         "in",
	OpNin:        "nin",
	OpMatch:      "~",
	OpNot:        "!",
	OpVal:        ":",
}

// GoString returns the internal name of Op o, for debugging
func (o Op) GoString() string {
	return opNames[o]
}

// String returns a source-level representation of Op o for diagnostics
func (o Op) String() string {
	return opText[o]
}

// Opcode returns the op value itself, so Op embedded in an operation satisfies the Expr interface.
func (o Op) Opcode() Op {
	return o
}

// IsLeaf returns true if o is a leaf operator
func (o Op) IsLeaf() bool {
	switch o {
	case OpID, OpString, OpInt, OpReal, OpRE, OpRoot, OpCurrent, OpWild, OpVal:
		return true
	default:
		return false
	}
}

// HasVal returns true if o is a leaf operator that carries a value.
func (o Op) HasVal() bool {
	switch o {
	case OpID, OpString, OpInt, OpReal, OpRE, OpVal:
		return true
	default:
		return false
	}
}

// precedence returns a binary operator's precedence, or -1 if it's not a binary operator.
// OpMatch (~, =~) is given the same precedence here as a relational operator,
// although some implementations put it below OpMul.
func (op Op) precedence() int {
	switch op {
	case OpOr:
		return 0
	case OpAnd:
		return 1
	case OpEQ, OpNE:
		return 2
	case OpLT, OpLE, OpGT, OpGE, OpMatch, OpIn, OpNin:
		return 3
	case OpAdd, OpSub:
		return 4
	case OpMul, OpDiv, OpMod:
		return 5
	default:
		return -1
	}
}

// associativity returns 1 for left-associative binary operators and 0 for right-associative binary operators
func (Op) associativity() int {
	return 1 // they are all left-associative at the moment
}
