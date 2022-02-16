// Copyright © 2021-22 Charles Forsyth (charles.forsyth@gmail.com)
// Usable under the terms in the file LICENSE.

/*
Package jsonpath provides a parser for JSONpaths, a syntax for expressing queries and locations in a JSON structure.

The JSONpath syntax is often defined by providing a set of sample paths.
Following https://github.com/dchester/jsonpath/, this package instead is based on a grammar,
Briefly, a JSONpath gives a dot-separated path through a JSON structure, with nested expressions providing dynamic values and filters.
For the detailed syntax, run
	go doc jsonpath/syntax

The semantics and built-in functions are generally those of https://danielaparker.github.io/JsonCons.Net/articles/JsonPath/Specification.html — a rare example of specifying JSONpath systematically instead of providing a few examples —  although this grammar is more restrictive (eg, as regards the content of a union expression). Some of Parker's extensions (eg, the parent operator) are also not provided.

ParsePath returns a Path that represents the JSONpath provided as text.
That result Path, and the Steps and Expr trees it contains, can then be evaluated against a subject JSON structure ("document").

ParseScriptExpression parses a string that contains only a script expression (not a path) and returns the Expr tree. It is not normally needed, because ParsePath will parse any script expressions in a path string, but might be useful for calculating using values in a JSON structure.

Path.Compile compiles a parsed Path into a Program for a small abstract machine.

Program.Run runs the program with a JSON structure as input ("the root document", or "$"), yielding the collection of JSON structures selected by the original path expression.
Several threads can Run the same Program simultaneously, since each Run gets its own abstract machine state.

(Components are still subject to change, since this is not yet an initial release.)
*/
package jsonpath
