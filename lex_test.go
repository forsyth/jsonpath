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
		[]string{"$", "tokNest", "tokID:book", "[", "(", "@", ".", "tokID:length", "-", "tokInt:1", ")", "]", "tokEOF"},
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
		f += ":" + lx.err.Error()
	} else if lx.tok.hasVal() {
		switch lx.tok {
		case tokInt:
			f += ":" + fmt.Sprint(lx.i())
		case tokID:
			f += ":" + lx.s()
		case tokString:
			f += ":" + fmt.Sprintf("%#v", lx.s())
		}
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
	// TO DO: switch to same scheme as path_test, with the expected output stream as plain text.
	building := testing.Verbose()
Samples:
	for i, sam := range samples {
		rdr := &rd{s: sam.s}
		ls := &lexState{lexer: lexer{r: rdr}}
		if building {
			fmt.Printf("%s ->", sam.s)
		}
		for j, expect := range sam.ops {
			lx := ls.lex()
			got := testForm(lx)
			if building {
				fmt.Printf(" %s", got)
			}
			if got != expect {
				if building {
					fmt.Print("\n")
				}
				t.Errorf("sample %d, token %d, got %s; expected %s", i+1, j+1, got, expect)
				// no point printing tokens, because the state can be messed up
				continue Samples
			}
			if lx.tok == tokError {
				if building {
					fmt.Print("\n")
				}
				continue Samples
			}
		}
		if rdr.look() != eof {
			t.Errorf("sample %d, reference stopped before EOF", i+1)
			if building {
				fmt.Print(" # ")
			}
			for {
				lx := ls.lex()
				if building {
					fmt.Printf(" %s", testForm(lx))
				}
				if lx.tok == tokEOF || lx.tok == tokError {
					break
				}
			}
		}
		if building {
			fmt.Print("\n")
		}
	}
}
