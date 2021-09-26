package JSONPath

import (
	"fmt"
	"testing"
)

//type el struct {
//	tok token
//	val interface{}
//}

//type lexOutput struct {
//	s   string
//	ops []el
//}

//var samples []lexOutput = []lexOutput{
//	lexOutput{"$", []el{{tok: '$'}, el{tok: tokEOF}}},
//	lexOutput{"$.store.book[(@.length-1)].title",
//		[]el{{tok: '$'}, {tok: '.'}, {tok: tokID, val: "store"}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: '('},
//			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "length"}, {tok: '-'}, {tok: tokInt, val: 1}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID}, {tok: tokEOF},
//		},
//	},
//	lexOutput{"$.store.book[?(@.price < 10)].title",
//		[]el{{tok: '$'}, {tok: '.'}, {tok: tokID}, {tok: '.'}, {tok: tokID}, {tok: '['}, {tok: tokFilter},
//			{tok: '@'}, {tok: '.'}, {tok: tokID}, {tok: '<'}, {tok: tokInt, val: 10}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
//		},
//	},
//	lexOutput{"$.['store'].book[?(@.price < 10)].title",
//		[]el{{tok: '$'}, {tok: '.'}, {tok: '['}, {tok: tokString, val: "store"}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: tokFilter},
//			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: '<'}, {tok: tokInt, val: 10}, {tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
//		},
//	},
//	lexOutput{"$..book[(@.length-1)]",
//		[]el{{tok: '$'}, {tok: tokNest}, {tok: tokID, val: "book"}, {tok: '['}, {tok: '('}, {tok: '@'}, {tok: '.'}, {tok: tokID, val: "length"}, {tok: '-'}, {tok: tokInt, val: 1}, {tok: ')'}, {tok: ']'}, {tok: tokEOF}},
//	},
//	lexOutput{"$.['store'].book[?(@.price >= 20 && @.price <= 50 || (  true \t))].title",
//		[]el{{tok: '$'}, {tok: '.'}, {tok: '['}, {tok: tokString, val: "store"}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "book"}, {tok: '['}, {tok: tokFilter},
//			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: tokGE}, {tok: tokInt, val: 20}, {tok: tokAnd},
//			{tok: '@'}, {tok: '.'}, {tok: tokID, val: "price"}, {tok: tokLE}, {tok: tokInt, val: 50}, {tok: tokOr}, {tok: '('}, {tok: tokID, val: "true"}, {tok: ')'},
//			{tok: ')'}, {tok: ']'}, {tok: '.'}, {tok: tokID, val: "title"}, {tok: tokEOF},
//		},
//	},
//}

var parseSamples []string = []string {
	"$",
	"$.store.book[?(@.price < 10)].title",
	"$['store'].book[?(@.price < 10)].title",
	"$..book[(@.length-1)]",
	"$['store'].book[?(@.price >= 20 && @.price <= 50 || (  true \t))].title",
}

//// keep enough state to handle nested script-expressions [nested ()]
//type lexState struct {
//	lexer
//	nestp int
//	expr  bool
//}

//func (ls *lexState) lex() lexeme {
//	if ls.expr {
//		lx := ls.lexExpr(false)
//		switch lx.tok {
//		case '(':
//			ls.nestp++
//		case ')':
//			if ls.nestp > 0 {
//				ls.nestp--
//			}
//			if ls.nestp == 0 {
//				ls.expr = false
//			}
//		}
//		return lx
//	}
//	lx := ls.lexPath()
//	if lx.tok == '(' || lx.tok == tokFilter {
//		ls.nestp++
//		ls.expr = true
//	}
//	return lx
//}

func TestPathParse(t *testing.T) {
	for _, sam := range parseSamples {
		fmt.Printf("%s -> ", sam)
		path, err := ParsePath(sam)
		if err != nil {
			fmt.Printf("!%s\n", err)
			continue
		}
		for _, el := range path {
			fmt.Printf(" %s", el)
		}
		fmt.Printf("\n")
	}
}

func (s *Step) String() string {
	doc := s.op.GoString()
	if len(s.args) > 0 {
		doc += "("
		for i, a := range s.args {
			if i > 0 {
				doc += ","
			}
			doc += fmt.Sprintf("%#v", a)
		}
		doc += ")"
	}
	return doc
}
