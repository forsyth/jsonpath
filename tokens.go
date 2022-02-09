package jsonpath

import "unicode/utf8"

// A token is an value in the lexical stream produced by the lexical analyser.
// Many tokens are represented directly by the rune value (typically in the ASCII range), eg '*', '(', '['.
// OpThers are compound tokens that represent an operator such as "<=", or
// an identifier, number or string, with an associated value.
// For simplicity, the stream of tokens is usable by path, filter and expression parsers,
// since inappropriate tokens in the sequence must anyway be detected by the parsers.
type token rune

const (
	tokError  token = utf8.MaxRune + iota // not a valid token in any grammar
	tokEOF                                // end of file
	tokID                                 // identifier
	tokString                             // single- or double-quoted string
	tokInt                                // integer
	tokNest                               // ..
	tokReal                               // real number (might be used in expressions)
	tokRE                                 // /re/, in expressions
	tokLE                                 // <=
	tokGE                                 // >=
	tokEQ                                 // ==
	tokNE                                 // !=
	tokFilter                             // ?(
	tokAnd                                // &&
	tokOr                                 // ||
	tokMatch                              // =~
	tokIn                                 // "in"
	tokNin                                // "nin"
)

// hasVal returns true if token t has an associated value
func (t token) hasVal() bool {
	switch t {
	case tokID, tokString, tokInt, tokReal, tokRE:
		return true
	default:
		return false
	}
}

var tokNames = map[token]string{
	tokError:  "tokError",
	tokEOF:    "tokEOF",
	tokID:     "tokID",
	tokString: "tokString",
	tokInt:    "tokInt",
	tokNest:   "tokNest",
	tokReal:   "tokReal",
	tokRE:     "tokRE",
	tokLE:     "tokLE",
	tokGE:     "tokGE",
	tokEQ:     "tokEQ",
	tokNE:     "tokNE",
	tokFilter: "tokFilter",
	tokAnd:    "tokAnd",
	tokOr:     "tokOr",
	tokMatch:  "tokMatch",
	tokIn:     "tokIn",
	tokNin:    "tokNin",
}

// GoString returns the internal name of a token (for debugging)
func (t token) GoString() string {
	if s := tokNames[t]; s != "" {
		return s
	}
	return string(t)
}

var tokText = map[token]string{
	tokError:  "(invalid token)",
	tokEOF:    "end of expression",
	tokID:     "identifier",
	tokString: "string literal",
	tokInt:    "integer literal",
	tokNest:   "..",
	tokReal:   "floating-point literal",
	tokRE:     "regular expression",
	tokLE:     "<=",
	tokGE:     ">=",
	tokEQ:     "==",
	tokNE:     "!=",
	tokFilter: "?(",
	tokAnd:    "&&",
	tokOr:     "||",
	tokMatch:  "=~",
	tokIn:     "in",
	tokNin:    "nin",
}

// String returns an readable form of a token for diagnostics
func (t token) String() string {
	if s := tokText[t]; s != "" {
		return s
	}
	return string(t)
}
