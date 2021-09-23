package JSONPath

import (
	"errors"
	"fmt"
)

type Val interface{}

type Path []*Step

type Step struct {
	op   Op
	args []Val
}

// path ::= "$" step*
// step ::= "." name | ".." name | "[" union-element ("," union-element)* "]"
// name ::= identifier | string | "*"
// union-element ::= array-index | array-slice | filter
// array-index ::= integer | expr
// array-slice ::= start? ":" end? (":" step?)?
// start ::= integer | expr
// end ::= integer | expr
// step ::= integer | expr
// expr ::= "(" script-expression ")"
// filter ::= "?(" script-expression ")"
// integer := "-"? [0-9]+
func ParsePath(s string) (Path, error) {
	rdr := &rd{s: s}
	err := expect(rdr, '$')
	if err != nil {
		return nil, err
	}
	path := []*Step{&Step{OpRoot, nil}}
	for {
		var step *Step
		tok, val, err := lexPath(rdr)
		switch tok {
		case tokError:
			return nil, err
		case '(':
			e, err := parseExpr(rdr)
			if err != nil {
				return nil, err
			}
			step = &Step{OpExp, []Val{e}}
		case tokFilter:
			e, err := parseExpr(rdr)
			if err != nil {
				return nil, err
			}
			step = &Step{OpFilter, []Val{e}}
		case '.':
			op, name, err := parsePathName(rdr)
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				step = &Step{OpWild, nil}
			} else {
				step = &Step{op, []Val{name}}
			}
		case tokNest:
			op, name, err := parsePathName(rdr)
			if err != nil {
				return nil, err
			}
			if op == OpWild {
				return nil, errors.New("..* not allowed") // or is it?
			}
			step = &Step{OpNest, []Val{name}}
		case '[':
			op, vals, err := parseVals(rdr)
			if err != nil {
				return nil, err
			}
			err = expect(rdr, ']')
			if err != nil {
				return nil, err
			}
			_ = op   // OpSlice, OpIndex, OpSelect, OpUnion
			_ = vals // need to inspect them to distinguish
		default:
			if tok.hasVal() {
				return nil, fmt.Errorf("unexpected token %v (%v)", tok, val)
			}
			return nil, fmt.Errorf("unexpected token %v", tok)
		}
		path = append(path, step)
	}
	return path, nil
}

func parseVals(r *rd) (Op, []Val, error) {
	return OpError, nil, errors.New("parseVals not done yet")
}

func parseExpr(r *rd) (Expr, error) {
	p := &parser{r: r}
	return p.parse()
}

// identifier, string or *
func parsePathName(r *rd) (Op, string, error) {
	tok, val, err := lexPath(r)
	if err != nil {
		return OpError, "", err
	}
	switch tok {
	case '*':
		return OpWild, "", nil
	case tokID:
		return OpId, val.(string), nil
	case tokString:
		return OpString, val.(string), nil
	default:
		return OpError, "", fmt.Errorf("unexpected %v at %s", tok, r.offset())
	}
}

func expect(r *rd, nt token) error {
	tok, _, err := lexPath(r)
	if err != nil {
		return err
	}
	if tok != nt {
		return fmt.Errorf("expected %q at %s, got %v", nt, r.offset(), tok)
	}
	return nil
}
