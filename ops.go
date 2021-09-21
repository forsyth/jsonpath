package JSONPath

// Op represents path expression leaf and expression operators, and filter and expression engine operators.
// Op values also are the tokens produced by the lexical analyser.
type Op int

const (
	Oerror  Op = iota // illegal token, deliberately the same as the zero value
	Oeof              // end of file
	Oid               // identifier
	Ostring           // single- or double-quoted string
	Oint              // integer
	Oreal             // real number (might be used in expressions)
	Ore               // /re/

	// path operators
	Oroot    // $
	Ocurrent // @
	Odot     // .
	Oselect  // [] when used for selection
	Oindex   // [] when used for indexing
	Oslice   // [lb: ub: step] slice operator
	Ounion   // [key1, key2 ...]
	Owild    // *
	Onest    // ..
	Ofilter  // ?(...)
	Oexp     // (...)

	// expression operators, in both filters and "expression engines"
	Olt    // <
	Ole    // <=
	Oeq    // = or ==
	One    // !=
	Oge    // >=
	Ogt    // >
	Oand   // &&
	Oor    // ||
	Omul   // *
	Odiv   // /
	Omod   // %
	Oneg   // unary -
	Oadd   // +
	Osub   // binary -
	Ocall  // function call id(args)
	Olist  // lists of things (eg, arguments)
	Oin    // "in"
	Onin   // "nin", not in
	Omatch // ~

	// used as tokens until the context is clear
	Olpar  // (
	Orpar  // )
	Obra   // [
	Oket   // ]
	Ocolon // :
	Ocomma // ,
	Ostar  // *
	Oslash // /
)

var opnames map[Op]string = map[Op]string{
	Oerror:   "Oerror",
	Oeof:     "Oeof",
	Oid:      "Oid",
	Ostring:  "Ostring",
	Oint:     "Oint",
	Oreal:    "Oreal",
	Ore:      "Ore",
	Oroot:    "Oroot",
	Ocurrent: "Ocurrent",
	Odot:     "Odot",
	Oselect:  "Oselect",
	Oindex:   "Oindex",
	Oslice:   "Oslice",
	Ounion:   "Ounion",
	Owild:    "Owild",
	Onest:    "Onest",
	Ofilter:  "Ofilter",
	Oexp:     "Oexp",
	Olt:      "Olt",
	Ole:      "Ole",
	Oeq:      "Oeq",
	One:      "One",
	Oge:      "Oge",
	Ogt:      "Ogt",
	Oand:     "Oand",
	Oor:      "Oor",
	Omul:     "Omul",
	Odiv:     "Odiv",
	Omod:     "Omod",
	Oneg:     "Oneg",
	Oadd:     "Oadd",
	Osub:     "Osub",
	Ocall:    "Ocall",
	Olist:    "Olist",
	Oin:      "Oin",
	Onin:     "Onin",
	Omatch:   "Omatch",
	Olpar:    "Olpar",
	Orpar:    "Orpar",
	Obra:     "Obra",
	Oket:     "Oket",
	Ocolon:   "Ocolon",
	Ocomma:   "Ocomma",
	Ostar:    "Ostar",
	Oslash:   "Oslash",
}

// String returns the textual representation of Op o
func (o Op) String() string {
	return opnames[o]
}
