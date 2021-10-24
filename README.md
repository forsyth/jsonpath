# jsonpath
Developing Go implementation of JSONpath

This is a Go implementation of lexing and parsing of JSONpath, with a conventional split between
lexical analysis (tokenising) and parsing (building from the token stream an abstract representation of a path sequence,
with nested expression trees).
That representation is converted into orders for a small abstract machine that evaluates a path expression given a root JSON value, yielding a collection of JSON values selected by the path expression.

It is currently in development, subject to change and not yet ready for use.

The path image is from Joey Genovese on Unsplash.

Paths have the following grammar:

	path ::= "$" step*
	step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
	member ::= "*" | identifier | expr | signed-integer
	subscript ::= subscript-expression | union-element ("," union-element)
	subscript-expression ::= "*" | expr | filter
	union-element ::=  array-index | string-literal | array-slice
	array-index ::= signed-integer
	array-slice ::= start? ":" end? (":" stride?)?
	start ::= signed-integer | expr
	end ::= signed-integer | expr
	stride ::= signed-integer | expr
	expr ::= "(" script-expression ")"
	filter ::= "?(" script-expression ")"
	step ::= ...  "[" subscript "]" ... | ".." "[" subscript "]"
	subscript ::= subscript-expression | union-element ("," union-element)
	subscript-expression ::= "*" | expr | filter
	union-element ::=  array-index | string-literal | array-slice
	array-index ::= signed-integer
	array-slice ::= start? ":" end? (":" step?)?
	member ::= "*" | identifier | expr | signed-integer
	expr ::= "(" script-expression ")"
	signed-integer ::= "-"? integer
	integer ::= [0-9]+

Script expressions (filters and calculations) share the same syntax:

	script-expression ::= e   // both filters and values share the same syntax
	e ::= primary | e binary-op e | unary-op e
	binary-op ::= "+" | "-" | "*" | "/" | "%" | "<" | ">" |
		">=" | "<=" | "==" | "!=" | "~" | "in" | "nin"  | "&&" | "||"
	unary-op ::= "-" | "!"
	unary ::= ("-" | "!")+ primary
	primary ::= primary1 ("(" e-list? ")" | "[" e "]" | "." identifier)*
	e-list ::= e ("," e)*
	primary1 ::= identifier | integer | real | string |
			"/" re "/" | "@" | "$" | "(" e ")" | "[" e-list? "]"
	re ::= <regular expression of some style, with \/ escaping the delimiting "/">
	real ::= integer "." integer? ("e" [+-]? integer)?

The semantics and built-in functions are generally those of https://danielaparker.github.io/JsonCons.Net/articles/JsonPath/Specification.html — a rare example of specifying JSONPath systematically instead of providing a few examples —  although this grammar is more restrictive (eg, as regards the content of a union expression).
