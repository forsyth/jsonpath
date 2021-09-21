package JSONPath

import "unicode"

// A token is an value in the lexical stream produced by the lexical analyser.
// Many tokens are represented directly by the rune value (typically in the ASCII range), eg '*', '(', '['.
// Others are compound tokens that represent an operator such as "<=", or
// an identifier, number or string, with an associated value.
// For simplicity, the stream of tokens is usable by path, filter and expression parsers,
// since inappropriate tokens in the sequence must anyway be detected by the parsers.
type token rune

const (
	tokError  token = unicode.MaxRune + iota // not a valid token in any grammar
	tokEOF                                   // end of file
	tokID                                    // identifier
	tokString                                // single- or double-quoted string
	tokInt                                   // integer
	tokNest                                  // ..
	tokReal                                  // real number (might be used in expressions)
	tokRE                                    // /re/, in expressions
	tokLE                                    // <=
	tokGE                                    // >=
	tokEq                                    // ==
	tokNE                                    // !=
	tokFilter                                // ?(
	tokAnd                                   // &&
	tokOr                                    // ||
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

// String returns a printable form of a token
func (t token) String() string {
	switch t {
	case tokError:
		return "tokError"
	case tokEOF:
		return "tokEOF"
	case tokID:
		return "tokID"
	case tokString:
		return "tokString"
	case tokInt:
		return "tokInt"
	case tokNest:
		return "tokNest"
	case tokReal:
		return "tokReal"
	case tokRE:
		return "tokRE"
	case tokLE:
		return "tokLE"
	case tokGE:
		return "tokGE"
	case tokEq:
		return "tokEq"
	case tokNE:
		return "tokNE"
	case tokFilter:
		return "tokFilter"
	case tokAnd:
		return "tokAnd"
	case tokOr:
		return "tokOr"
	default:
		return string(t)
	}
}
