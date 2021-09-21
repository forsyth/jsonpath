package JSONPath

import (
	"fmt"
	"testing"
)

type el struct {
	tok  token
	val interface{}
}

type lexOutput struct {
	s   string
	ops []el
}

var samples []lexOutput = []lexOutput{
	lexOutput{"$", []el{{tok: '$'}, el{tok: tokEOF}}},
	lexOutput{"$.store.book[(@.length-1)].title",
		[]el{el{tok: '$'}, el{tok: '.'}, el{tok: tokID, val: "store"}, el{tok: '.'}, el{tok: tokID, val: "book"}, el{tok: '['}, el{tok: '('},
		      el{tok: '@'}, el{tok: '.'}, el{tok: tokID, val: "length"}, el{tok: '-'}, el{tok: tokInt, val: 1}, el{tok: ')'}, el{tok: ']'}, el{tok: '.'}, el{tok: tokID}, el{tok: tokEOF},
		},
	},
	lexOutput{"$.store.book[?(@.price < 10)].title",
		[]el{el{tok: '$'}, el{tok: '.'}, el{tok: tokID}, el{tok: '.'}, el{tok: tokID}, el{tok: '['}, el{tok: tokFilter},
		      el{tok: '@'}, el{tok: '.'}, el{tok: tokID}, el{tok: '<'}, el{tok: tokInt, val: 10}, el{tok: ')'}, el{tok: ']'}, el{tok: '.'}, el{tok: tokID, val: "title"}, el{tok: tokEOF},
		},
	},
	lexOutput{"$.['store'].book[?(@.price < 10)].title",
		[]el{el{tok: '$'}, el{tok: '.'}, el{tok: '['}, el{tok: tokString, val: "store"}, el{tok: ']'},  el{tok: '.'}, el{tok: tokID, val: "book"}, el{tok: '['}, el{tok: tokFilter},
		      el{tok: '@'}, el{tok: '.'}, el{tok: tokID, val: "price"}, el{tok: '<'}, el{tok: tokInt, val: 10}, el{tok: ')'}, el{tok: ']'}, el{tok: '.'}, el{tok: tokID, val: "title"}, el{tok: tokEOF},
		},
	},
	lexOutput{"$..book[(@.length-1)]",
		[]el{el{tok: '$'},  el{tok: tokNest},  el{tok: tokID, val:"book"}, el{tok: '['},  el{tok: '('},  el{tok: '@'},  el{tok: '.'},  el{tok: tokID, val: "length"}, el{tok: '-'},  el{tok: tokInt, val: 1},  el{tok: ')'},  el{tok: ']'},  el{tok: tokEOF}, 
		},
	},
	lexOutput{"$.['store'].book[?(@.price >= 20 && @.price <= 50 || true)].title",
		[]el{el{tok: '$'}, el{tok: '.'}, el{tok: '['}, el{tok: tokString, val: "store"}, el{tok: ']'},  el{tok: '.'}, el{tok: tokID, val: "book"}, el{tok: '['}, el{tok: tokFilter},
		      el{tok: '@'}, el{tok: '.'}, el{tok: tokID, val: "price"}, el{tok: tokGE}, el{tok: tokInt, val: 20}, el{tok: tokAnd},
			el{tok: '@'}, el{tok: '.'}, el{tok: tokID, val: "price"}, el{tok: tokLE}, el{tok: tokInt, val: 50}, el{tok: tokOr}, el{tok: tokID},
			el{tok: ')'}, el{tok: ']'}, el{tok: '.'}, el{tok: tokID, val: "title"}, el{tok: tokEOF},
		},
	},
}

type lexState struct {
	r	*rd
	nestp	int
	expr	bool
}

func (ls *lexState) lex() (token, interface{}, error) {
	if ls.expr {
		tok, v, err := lexExpr(ls.r, false)
		switch tok {
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
		return tok, v, err
	}
	tok, v, err := lexPath(ls.r)
	if tok == '(' || tok == tokFilter {
		ls.nestp++
		ls.expr = true
	}
	return tok, v, err
}

func TestLex(t *testing.T) {
	for i, sam := range samples {
		rdr := &rd{sam.s, 0}
		ls := &lexState{rdr, 0, false}
		fmt.Printf("%s -> ", sam.s)
		for j, el := range sam.ops {
			tok, val, err := ls.lex()
			print(tok, val, err)
			if tok != el.tok || tok != tokError && err != nil {
				t.Errorf("sample %d el %d, got %v (%#v %v) expected %v (%#v)", i, j, tok, val, err, el.tok, el.val)
				break
			}
		}
		fmt.Printf("\n")
		if rdr.look() != eof {
			t.Errorf("sample %d, not reached tokEOF", i)
			for {
				tok, val, err := ls.lex()
				print(tok, val, err)
				if tok == tokEOF || tok == tokError {
					break
				}
			}
			fmt.Printf("\n")
		}
	}
}

func print(tok token, val interface{}, err error) {
	fmt.Printf(" %v", tok)
	if tok.hasVal() {
		fmt.Printf("[%#v]", val)
	}
	if err != nil {
		fmt.Printf("!%s", err)
	}
}
