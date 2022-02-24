// Copyright © 2021-22 Charles Forsyth (charles.forsyth@gmail.com)
// Usable under the terms in the file LICENSE.

/*
The JSONpath syntax is typically defined by providing a set of sample paths.
Following https://github.com/dchester/jsonpath/, this package instead is based on a grammar.

The grammar has two distinct parts, one for the paths (the outermost construction),
and one for expressions and filter expressions that can appear in a path ("script-expressions").
The two parts have different structures.

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
	e ::= primary | e binary-op e
	binary-op ::= "+" | "-" | "*" | "/" | "%" | "<" | ">" | ">=" | "<=" | "==" | "!=" | "=~" | "in" | "nin"  | "&&" | "||"
	unary-op ::= "-" | "!"
	primary ::= primary1 ("(" e-list? ")" | "[" e "]" | "." identifier)*
	e-list ::= e ("," e)*
	primary1 ::= identifier | integer | real | string | "/" re "/" | "@" | "$" | "(" e ")" | "[" e-list? "]" | unary-op primary1
	re ::= <regular expression of some style, with \/ escaping the delimiting "/">
	real ::= integer "." integer? ("e" [+-]? integer)?

The semantics and built-in functions are generally those of https://danielaparker.github.io/JsonCons.Net/articles/JsonPath/Specification.html — a rare example of specifying JSONpath systematically instead of providing a few examples —  although the grammar above is more restrictive (eg, as regards the content of a union expression). Some of Parker's extensions (eg, the parent operator) are also not provided.

JSONpath expressions were originally described by https://goessner.net/articles/JsonPath/index.html by
analogy with XPath for XML.
*/
package syntax
