/*
Package mach implements a small abstract machine for path expressions, based on the representation produced by the sibling package paths.

mach.Compile compiles a parsed paths.Path into a Program for a small abstract machine.

Program.Run runs the program with a JSON structure as input ("the root document", or "$"), yielding the collection of JSON structures selected by the original path expression.
Several threads can Run the same Program simultaneously, since each Run gets its own abstract machine state.

The semantics and built-in functions are generally those of https://danielaparker.github.io/JsonCons.Net/articles/JsonPath/Specification.html — a rare example of specifying JSONpath systematically instead of providing a few examples —  although the grammar above is more restrictive (eg, as regards the content of a union expression). Some of Parker's extensions (eg, the parent operator) are also not provided.
*/
package mach
