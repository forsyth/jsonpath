package JSONPath

import (
	"fmt"
	"testing"
)

type el struct {
	tok token
	val interface{}
}

type lexOutput struct {
	s   string
	ops []el
}

var samples []lexOutput = []lexOutput{
	lexOutput{"$", []el{{tok: '$'}, el{tok: tokEOF}}},
	lexOutput{"$[-99]", []el{{tok: '$'}, el{tok: '['}, el{tok: tokInt, val: -99}, el{tok: ']'}, el{tok: tokEOF}}},
	lexOutput{"$[-9223372036854775807]", []el{{tok: '$'}, el{tok: '['}, el{tok: tokError, val: "overflow of negative integer literal"}, el{tok: ']'}, el{tok: tokEOF}}},
	lexOutput{"$.store.book[(@.length-1)].title",
		[]el{{tok: '$'}, {tok: '.'}, {tok: tokID, val: "store"}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: '('},
			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "length"}, {tok: '-'}, {tok: tokInt, val: 1}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID}, {tok: tokEOF},
		},
	},
	lexOutput{"$.store.book[?(@.price < 10)].title",
		[]el{{tok: '$'}, {tok: '.'}, {tok: tokID}, {tok: '.'}, {tok: tokID}, {tok: '['}, {tok: tokFilter},
			{tok: '@'}, {tok: '.'}, {tok: tokID}, {tok: '<'}, {tok: tokInt, val: 10}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
		},
	},
	lexOutput{"$['store'].book[?(@.price < 10)].title",
		[]el{{tok: '$'}, {tok: '['}, {tok: tokString, val: "store"}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: tokFilter},
			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: '<'}, {tok: tokInt, val: 10}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
		},
	},
	lexOutput{"$..book[(@.length-1)]",
		[]el{{tok: '$'}, {tok: tokNest}, {tok: tokID, val: "book"}, {tok: '['}, {tok: '('}, {tok: '@'}, {tok: '.'}, {tok: tokID, val: "length"}, {tok: '-'}, {tok: tokInt, val: 1}, {tok: ')'}, {tok: ']'}, {tok: tokEOF}},
	},
	lexOutput{"$['store'].book[?(@.price >= 20 && @.price <= 50 || (  true \t))].title",
		[]el{{tok: '$'}, {tok: '['}, {tok: tokString, val: "store"}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: tokFilter},
			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: tokGE}, {tok: tokInt, val: 20}, {tok: tokAnd},
			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: tokLE}, {tok: tokInt, val: 50}, {tok: tokOr}, {tok: '('}, {tok: tokID, val: "true"}, {tok: ')'},
			{tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
		},
	},
	lexOutput{"$[':@.\"$,*\\'\\\\']",
		[]el{{tok: '$'}, {tok: '['}, {tok: tokString, val: ":@.\"$,*'\\"}, {tok: ']'}, {tok: tokEOF}},
	},
	lexOutput{"$[':@.\"$,*\\\\'\\\\\\\\']",
		[]el{{tok: '$'}, {tok: '['}, {tok: tokString, val: ":@.\"$,*\\"}, {tok: tokError, val: "unexpected character '\\\\' at offset 13"}},
	},
}

// keep enough state to handle nested script-expressions [nested ()]
type lexState struct {
	lexer
	nestp int // nesting count for ()
}

// lex switches between the path lexer and expression lexer, at the outermost ( or ?( and back at the closing )
func (ls *lexState) lex() lexeme {
	if ls.nestp > 0 {
		lx := ls.lexExpr()
		switch lx.tok {
		case '(':
			ls.nestp++
		case ')':
			if ls.nestp > 0 {
				ls.nestp--
			}
		}
		return lx
	}
	lx := ls.lexPath()
	if lx.tok == '(' || lx.tok == tokFilter {
		ls.nestp++
	}
	return lx
}

func TestLex(t *testing.T) {
Samples:
	for i, sam := range samples {
		rdr := &rd{s: sam.s}
		ls := &lexState{lexer: lexer{r: rdr}}
		fmt.Printf("%s -> ", sam.s)
		for j, el := range sam.ops {
			lx := ls.lex()
			fmt.Printf(" %#v", lx)
			if el.tok == tokError && el.val != nil {
				if lx.err == nil {
					t.Errorf("sample %d el %d, expected error (%s) got nil", i, j, el.val)
				} else if lx.err.Error() != el.val {
					t.Errorf("sample %d el %d, expected error (%s) got (%s)", i, j, el.val, lx.err)
				}
				fmt.Printf("\n")
				continue Samples
			}
			if lx.tok != el.tok || lx.tok != tokError && lx.err != nil {
				t.Errorf("sample %d el %d, got %v (%#v %v) expected %v (%#v)", i, j, lx.tok, lx.val, lx.err, el.tok, el.val)
				break
			}
		}
		fmt.Printf("\n")
		if rdr.look() != eof {
			t.Errorf("sample %d, not reached tokEOF", i)
			for {
				lx := ls.lex()
				fmt.Printf(" %#v", lx)
				if lx.tok == tokEOF || lx.tok == tokError {
					break
				}
			}
			fmt.Printf("\n")
		}
	}
}
