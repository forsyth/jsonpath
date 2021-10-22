package JSONPath

// Op represents path expression leaf and expression operators, and filter and expression engine operators.
//
// The "." operator in the grammar is transformed to different Ops that identify
// the following component (eg, OpMember, OpSelect).
// The ".." operator is similarly transformed, to distinguish the various
// cases of "subscript" in ".." "[" subscript "]".
// OpFor marks the start of an iteration over the current output set,
// refining it by applying the OpFilter expression to each set member.
// OpNest similarly marks the start of an iteration over the results of the
// recursive walk specified by "..", and sets up machine state to consider
// each value from the substructure as it is produced by the walk.

type Op int

const (
	OpError  Op = iota // illegal token, deliberately the same as the zero value
	OpEof              // end of file
	OpID               // identifier
	OpString           // single- or double-quoted string
	OpInt              // integer
	OpReal             // real number (might be used in expressions)
	OpRE               // /re/
	OpBounds           // [lb: ub: stride]

	// path operators
	OpMember // . used for path selection (single int, key or expr)
	OpSelect // [] for path selection (single int, string, expr or filter
	OpUnion  // [union-element, union-element ...]
	OpWild   // *
	OpFilter // ?(...)
	OpExp    // (...)

	// path iteration operators
	OpFor  // start of OpFilter sequence, selecting on output candidates
	OpNest // start of OpNest* sequence, selecting on dot
	OpRep  // repeat sequence if values left

	// path nest operators
	OpNestMember // .. member
	OpNestSelect // .. [subscript]
	OpNestUnion  // .. [key1, key2, ...]
	OpNestWild   // .. [*]
	OpNestFilter // .. [?(expr)]

	// expression operators, in both filters and "expression engines"
	OpRoot    // $ (use root as operand)
	OpCurrent // @ (use current candidate as operand)
	OpDot     // . field selection (in an expression)
	OpIndex   // [] indexing an array
	OpSlice   // [lb: ub: stride] slice operator on array value
	OpLT      // <
	OpLE      // <=
	OpEQ      // = or ==
	OpNE      // !=
	OpGE      // >=
	OpGT      // >
	OpAnd     // &&
	OpOr      // ||
	OpMul     // *
	OpDiv     // /
	OpMod     // %
	OpNeg     // unary -
	OpAdd     // +
	OpSub     // binary -
	OpCall    // function call id(args)
	OpArray   // [e-list]
	OpIn      // "in"
	OpNin     // "nin", not in
	OpMatch   // ~= (why not just ~)
	OpNot     // unary !
)

var opNames map[Op]string = map[Op]string{
	OpError:      "OpError",
	OpEof:        "OpEof",
	OpID:         "OpID",
	OpString:     "OpString",
	OpInt:        "OpInt",
	OpReal:       "OpReal",
	OpRE:         "OpRE",
	OpBounds:     "OpBounds",
	OpRoot:       "OpRoot",
	OpCurrent:    "OpCurrent",
	OpDot:        "OpDot",
	OpSelect:     "OpSelect",
	OpMember:     "OpMember",
	OpIndex:      "OpIndex",
	OpSlice:      "OpSlice",
	OpUnion:      "OpUnion",
	OpWild:       "OpWild",
	OpFilter:     "OpFilter",
	OpExp:        "OpExp",
	OpFor:        "OpFor",
	OpRep:        "OpRep",
	OpNest:       "OpNest",
	OpNestMember: "OpNestMember",
	OpNestSelect: "OpNestSelect",
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
}

var opText map[Op]string = map[Op]string{
	OpError:      "(error)",
	OpEof:        "(eof)",
	OpID:         "identifier",
	OpString:     "string",
	OpInt:        "integer",
	OpReal:       "real number",
	OpRE:         "regular expression",
	OpBounds:     "[lb:ub:stride]",
	OpRoot:       "$",
	OpCurrent:    "@",
	OpDot:        ".",
	OpSelect:     "[]selection",
	OpMember:     ". selection",
	OpIndex:      "[]index",
	OpSlice:      "[]slice",
	OpUnion:      "[]union",
	OpWild:       "*",
	OpFilter:     "?(filter)",
	OpExp:        "(exp)",
	OpFor:        "loop start",
	OpRep:        "loop end",
	OpNest:       "..",
	OpNestMember: "..member",
	OpNestSelect: "..[]selection",
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
}

// GoString returns the internal name of Op o, for debugging.
func (o Op) GoString() string {
	return opNames[o]
}

// String returns a source-level representation of Op o for diagnostics.
func (o Op) String() string {
	return opText[o]
}

// Opcode returns the op value itself, so Op embedded in an operation satisfies the Expr interface.
func (o Op) Opcode() Op {
	return o
}

// IsLeaf returns true if o is a leaf operator.
func (o Op) IsLeaf() bool {
	switch o {
	case OpID, OpString, OpInt, OpReal, OpRE, OpRoot, OpCurrent, OpWild, OpBounds:
		return true
	default:
		return false
	}
}

// HasVal returns true if o is a leaf operator that carries a value.
func (o Op) HasVal() bool {
	switch o {
	case OpID, OpString, OpInt, OpReal, OpRE, OpBounds:
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
