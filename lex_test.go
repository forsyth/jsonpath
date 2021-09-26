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
}

// keep enough state to handle nested script-expressions [nested ()]
type lexState struct {
	lexer
	nestp int
	expr  bool
}

func (ls *lexState) lex() lexeme {
	if ls.expr {
		lx := ls.lexExpr(false)
		switch lx.tok {
		case '(':
			ls.nestp++
		case ')':
			if ls.nestp > 0 {
				ls.nestp--
			}
			if ls.nestp == 0 {
				ls.expr = false
			}
		}
		return lx
	}
	lx := ls.lexPath()
	if lx.tok == '(' || lx.tok == tokFilter {
		ls.nestp++
		ls.expr = true
	}
	return lx
}

func TestLex(t *testing.T) {
	for i, sam := range samples {
		rdr := &rd{sam.s, 0}
		ls := &lexState{lexer: lexer{r: rdr}}
		fmt.Printf("%s -> ", sam.s)
		for j, el := range sam.ops {
			lx := ls.lex()
			fmt.Printf(" %#v", lx)
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
