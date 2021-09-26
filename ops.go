package JSONPath

// Op represents path expression leaf and expression operators, and filter and expression engine operators.
// Op values also are the tokens produced by the lexical analyser.
type Op int

const (
	OpError  Op = iota // illegal token, deliberately the same as the zero value
	OpEof              // end of file
	OpId               // identifier
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
	OpSlice   // [lb: ub: step] slice operator
	OpUnion   // [key1, key2 ...]
	OpWild    // *
	OpNest    // ..
	OpFilter  // ?(...)
	OpExp     // (...)

	// expression operators, in both filters and "expression engines"
	OpLt    // <
	OpLe    // <=
	OpEq    // = or ==
	OpNe    // !=
	OpGe    // >=
	OpGt    // >
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
)

var opNames map[Op]string = map[Op]string{
	OpError:   "OpError",
	OpEof:     "OpEof",
	OpId:      "OpId",
	OpString:  "OpString",
	OpInt:     "OpInt",
	OpReal:    "OpReal",
	OpRE:      "OpRE",
	OpRoot:    "OpRoot",
	OpCurrent: "OpCurrent",
	OpDot:     "OpDot",
	OpSelect:  "OpSelect",
	OpIndex:   "OpIndex",
	OpSlice:   "OpSlice",
	OpUnion:   "OpUnion",
	OpWild:    "OpWild",
	OpNest:    "OpNest",
	OpFilter:  "OpFilter",
	OpExp:     "OpExp",
	OpLt:      "OpLt",
	OpLe:      "OpLe",
	OpEq:      "OpEq",
	OpNe:      "OpNe",
	OpGe:      "OpGe",
	OpGt:      "OpGt",
	OpAnd:     "OpAnd",
	OpOr:      "OpOr",
	OpMul:     "OpMul",
	OpDiv:     "OpDiv",
	OpMod:     "OpMod",
	OpNeg:     "OpNeg",
	OpAdd:     "OpAdd",
	OpSub:     "OpSub",
	OpCall:    "OpCall",
	OpArray:   "OpArray",
	OpIn:      "OpIn",
	OpNin:     "OpNin",
	OpMatch:   "OpMatch",
	OpNot:     "OpNot",
}

var opText map[Op]string = map[Op]string{
	OpError:   "(error)",
	OpEof:     "(eof)",
	OpId:      "identifier",
	OpString:  "string",
	OpInt:     "integer",
	OpReal:    "real number",
	OpRE:      "regular expression",
	OpRoot:    "$",
	OpCurrent: "@",
	OpDot:     ".",
	OpSelect:  "[]selection",
	OpIndex:   "[]index",
	OpSlice:   "[]slice",
	OpUnion:   "[]union",
	OpWild:    "*",
	OpNest:    "..",
	OpFilter:  "?(filter)",
	OpExp:     "(exp)",
	OpLt:      "<",
	OpLe:      "<=",
	OpEq:      "==",
	OpNe:      "!=",
	OpGe:      ">=",
	OpGt:      ">",
	OpAnd:     "&&",
	OpOr:      "||",
	OpMul:     "*",
	OpDiv:     "/",
	OpMod:     "%",
	OpNeg:     "unary -",
	OpAdd:     "+",
	OpSub:     "-",
	OpCall:    "function call",
	OpArray:   "array value",
	OpIn:      "in",
	OpNin:     "nin",
	OpMatch:   "~",
	OpNot:     "!",
}

// GoString returns the textual representation of Op o, for debugging
func (o Op) GoString() string {
	return opNames[o]
}

// String returns a readable representation of Op o for diagnostics
func (o Op) String() string {
	return opText[o]
}

// IsLeaf returns true if o is a leaf operator
func (o Op) IsLeaf() bool {
	switch o {
	case OpId, OpString, OpInt, OpReal, OpRE, OpRoot, OpCurrent, OpWild:
		return true
	default:
		return false
	}
}
