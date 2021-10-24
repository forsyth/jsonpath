// Copyright © 2021 Charles Forsyth (charles.forsyth@gmail.com)
// Usable under the terms in the file LICENSE.

/*
Package JSONPath provides a parser for JSONpaths, a syntax for expressing queries and locations in a JSON structure.

The JSONpath syntax is typically defined by providing a set of sample paths.
Following https://github.com/dchester/jsonpath/, this package instead is based on a grammar,
and internally has a conventional lexical analyser and parser for the grammar.

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

The semantics and built-in functions are generally those of https://danielaparker.github.io/JsonCons.Net/articles/JsonPath/Specification.html — a rare example of specifying JSONPath systematically instead of providing a few examples —  although this grammar is more restrictive (eg, as regards the content of a union expression). Some of its extensions (eg, the parent operator) are also not provided.

Paths are represented by a Path type, which is just a sequence of Steps, each with some parameter values (Val). Expressions are represented by a type Expr, which is an expression tree.

ParsePath returns a Path that represents the JSONpath provided as text.
That result Path, and the Steps and Expr trees it contains, can then be evaluated against a subject JSON structure ("document").

ParseScriptExpression parses a string that contains only a script expression (not a path) and returns the Expr tree. It is not normally needed, because ParsePath will parse any script expressions in a path string, but might be useful for calculating using values in a JSON structure.

Path.Compile compiles a parsed Path into a Program for a small abstract machine.

Program.Run runs the program with a JSON structure as input ("the root document", or "$"), yielding the collection of JSON structures selected by the original path expression.
Several threads can Run the same Program simultaneously, since each Run gets its own abstract machine state.

(The evaluator is almost complete, but it and the Path, Step and Expr structures are still subject to change, since this project is not yet an initial release.
No issues yet, please!)
*/
package JSONPath
