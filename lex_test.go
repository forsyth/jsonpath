package JSONPath

import (
	"fmt"
	"testing"
)

type lexOutput struct {
	s   string
	ops []string
}

var samples []lexOutput = []lexOutput{
	lexOutput{"$", []string{"$", "tokEOF"}},
	lexOutput{"$[-99]",
		[]string{
			"$", "[", "tokInt:-99", "]", "tokEOF",
		},
	},
	lexOutput{"$[-9223372036854775807]", 
		[]string{
			"$", "[", "tokError:overflow of negative integer literal", "]", "tokEOF",
		},
	},
	lexOutput{"$.store.book[(@.length-1)].title",
		[]string{
			"$", ".", "tokID:store", ".", "tokID:book", "[", "(", "@", ".", "tokID:length", "-", "tokInt:1", ")", "]", ".", "tokID:title", "tokEOF",
		},
	},
	lexOutput{"$.store.book[?(@.price < 10)].title",
		[]string{
			"$", ".", "tokID:store", ".", "tokID:book", "[", "tokFilter", "@", ".", "tokID:price", "<", "tokInt:10", ")", "]", ".", "tokID:title", "tokEOF",
		},
	},
	lexOutput{"$['store'].book[?(@.price < 10)].title",
		[]string{
			"$", "[", "tokString:\"store\"", "]", ".", "tokID:book", "[", "tokFilter", "@", ".", "tokID:price", "<", "tokInt:10", ")", "]", ".", "tokID:title", "tokEOF",
		},
	},
	lexOutput{"$..book[(@.length-1)]",
		[]string{"$", "tokNest", "tokID:book", "[", "(", "@", ".", "tokID:length", "-", "tokInt:1", ")", "]", "tokEOF",},
	},
	lexOutput{"$['store'].book[?(@.price >= 20 && @.price <= 50 || (  true \t))].title",
		[]string{"$", "[", "tokString:\"store\"", "]", ".", "tokID:book", "[", "tokFilter", "@", ".", "tokID:price", "tokGE", "tokInt:20", "tokAnd",
			"@", ".", "tokID:price", "tokLE", "tokInt:50", "tokOr", "(", "tokID:true", ")", ")", "]", ".", "tokID:title", "tokEOF",
		},
	},
	lexOutput{"$[':@.\"$,*\\'\\\\']",
		[]string{"$", "[", "tokString:\":@.\\\"$,*'\\\\\"", "]", "tokEOF"},
	},
	lexOutput{"$[':@.\"$,*\\\\'\\\\\\\\']",
		[]string{"$", "[", "tokString:\":@.\\\"$,*\\\\\"", "tokError:unexpected character '\\\\' at offset 13"},
	},
}

func testForm(lx lexeme) string {
	f := lx.tok.GoString()
	if lx.tok == tokError {
		f += ":"+lx.err.Error()
	} else if lx.tok.hasVal() {
		f += ":"+lx.val.String()
	}
	return f
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
		fmt.Printf("%s ->", sam.s)
		for j, expect := range sam.ops {
			lx := ls.lex()
			got := testForm(lx)
			fmt.Printf(" %s", got)
			if got != expect {
				fmt.Print("\n")
				t.Errorf("sample %d, token %d, expected %s got %s", i+1, j+1, expect, got)
				// no point printing tokens, because the state can be messed up
				continue Samples
			}
			if lx.tok == tokError {
				continue Samples
			}
		}
		if rdr.look() != eof {
			t.Errorf("sample %d, reference stopped before EOF", i+1)
			fmt.Print(" # ")
			for {
				lx := ls.lex()
				fmt.Printf(" %s", testForm(lx))
				if lx.tok == tokEOF || lx.tok == tokError {
					break
				}
			}
		}
		fmt.Print("\n")
	}
}
