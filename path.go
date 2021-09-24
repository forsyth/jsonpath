package JSONPath

import (
	"errors"
	"fmt"
	"math"
)

type Val interface{}

type Path []*Step

type Step struct {
	op   Op
	args []Val
}

// started with the IETF drafts, but reverted to a grammar adapted from https://github.com/dchester/jsonpath/blob/master/lib/grammar.js
// path ::= "$" step*
// step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
// member ::= "*" | identifier | expr | integer
// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= integer
// array-slice ::= start? ":" end? (":" step?)?
// start ::= integer | expr
// end ::= integer | expr
// step ::= integer | expr
// expr ::= "(" script-expression ")"
// filter ::= "?(" script-expression ")"
// integer := "-"? [0-9]+

// ParsePath returns the parsed form of the path expression in s, or an error.
func ParsePath(s string) (Path, error) {
	rdr := &rd{s: s}
	err := expect(rdr, '$')
	if err != nil {
		return nil, err
	}
	path := []*Step{&Step{OpRoot, nil}}
	for {
		lx := lexPath(rdr)
		switch lx.tok {
		case tokEOF:
			return path, nil
		case tokError:
			return nil, lx.err
		case '.':
			op, name, err := parseMember(rdr)
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				path = append(path, &Step{OpWild, nil})
			} else {
				path = append(path, &Step{op, []Val{name}})
			}
		case tokNest:
//			if p.look() == '[' {
//				// ".." "[" subscript "]"
//				p.get()
//				sub, err := parseBrackets(rdr)
//				if err != nil {
//					return nil, err
//				}
//				path = append(path, &Step{OpNest, sub})
//				break
//			}
			op, name, err := parseMember(rdr)
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				return nil, errors.New("..* not allowed") // or is it?
			}
			path = append(path, &Step{OpNest, []Val{name}})
		case '[':
			sub, err := parseBrackets(rdr)
			if err != nil {
				return nil, err
			}
			path = append(path, sub)
		default:
			// TO DO: make lexeme printable
			if lx.tok.hasVal() {
				return nil, fmt.Errorf("unexpected token %v (%v)", lx.tok, lx.val)
			}
			return nil, fmt.Errorf("unexpected token %v", lx.tok)
		}
	}
}

// parse bracketed subscript in
// step ::= ...  "[" subscript "]" ... | ".." "[" subscript "]"
func parseBrackets(rdr *rd) (*Step, error) {
	op, vals, err := parseSubscript(rdr)
	if err != nil {
		return nil, err
	}
	err = expect(rdr, ']')
	if err != nil {
		return nil, err
	}
	// op= OpSlice, OpIndex, OpSelect, OpUnion
	// vals structure distinguishes them
	return &Step{op, []Val{vals}}, nil
}

// subscript ::= subscript-expression | union-element ("," union-element)
// subscript-expression ::= "*" | expr | filter
// union-element ::=  array-index | string-literal | array-slice   // could include identifier?
// array-index ::= integer
// array-slice ::= start? ":" end? (":" step?)?
func parseSubscript(rdr *rd) (Op, Val, error) {
	union := []*Step{}	// if it's the sequence of union-element
	for {
		lx := lexPath(rdr)
		switch lx.tok {

		case tokError:
			return OpError, "", lx.err
		case tokEOF:
			return OpError, "", unexpectedEOF(rdr)

		// subscript-expression
		case '*':
			return OpWild, nil, nil
		case '(':
			e, err := parseExpr(rdr)
			// need to lookahead for ":", the "expr" case of "start" (ie, it's a slice)
			return OpExp, e, err
		case tokFilter:
			e, err := parseExpr(rdr)
			return OpFilter, e, err

		// union-element ("," union-element)
		// case '-': // TO DO: signed integer; easier to add it to lexPath
		case tokInt:
			// integer or start element of slice
			// need to lookahead for ":" (OpIndex vs OpSlice)
			union = append(union, &Step{OpIndex, []Val{lx.val}})
		case tokString:
			// string-literal
			union = append(union, &Step{OpString, []Val{lx.val}})
		case tokID:
			// treat same as string-literal
			union = append(union, &Step{OpId, []Val{lx.val}})
		// error
		default:
			return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, rdr.offset())
		}
		lx = lexPath(rdr)
		if lx.tok != ',' {
			if lx.tok != ']' {
				// TO DO: unclosed bracket
			}
			return OpUnion, union, nil
		}
	}
}

// expr ::= "(" script-expression ")"
func parseExpr(r *rd) (Expr, error) {
	p := &parser{r: r}
	return p.parse()
}

// member ::= "*" | identifier | expr | integer
func parseMember(rdr *rd) (Op, Val, error) {
	lx := lexPath(rdr)
	if lx.err != nil {
		return OpError, "", lx.err
	}
	switch lx.tok {
	case '*':
		return OpWild, lx.val, nil
	case tokID:
		return OpId, lx.val, nil
	case tokString:
		return OpString, lx.val, nil
	case tokInt:
		return OpInt, lx.val, nil
	case '-':
		lx = lexPath(rdr)
		if lx.err != nil {
			return OpError, "", lx.err
		}
		if lx.tok != tokInt {
			return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, rdr.offset())
		}
		n := lx.val.(int64)
		if n == math.MaxInt64 {
			return OpError, "", fmt.Errorf("overflow with negative literal at %s", rdr.offset())
		}
		return OpInt, -n, nil
	case '(':
		e, err := parseExpr(rdr)
		if err != nil {
			return OpError, "", err
		}
		return OpExp, e, nil
	default:
		return OpError, "", fmt.Errorf("unexpected %v at %s", lx.tok, rdr.offset())
	}
}

func expect(r *rd, nt token) error {
	lx := lexPath(r)
	if lx.err != nil {
		return lx.err
	}
	if lx.tok != nt {
		return fmt.Errorf("expected %q at %s, got %v", nt, r.offset(), lx.tok)
	}
	return nil
}

func unexpectedEOF(r *rd) error {
	return fmt.Errorf("unexpected EOF at %s", r.offset())
}
