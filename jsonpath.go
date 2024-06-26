// Copyright © 2021-22 Charles Forsyth (charles.forsyth@gmail.com)
// Usable under the terms in the file LICENSE.

// Package jsonpath provides a parser and evaluator for JSONpaths, a syntax for expressing queries and locations in a JSON structure.
// Given a JSON value ("document") a path expression selects subcomponents and returns a list of the selected values.
//
// The JSONpath syntax is often defined by providing a set of sample paths.
// This package instead is based on a grammar, parsed and interpreted in a conventional way.
//
// Briefly, a JSONpath gives a dot-separated path through a JSON structure, with nested expressions providing dynamic values and filters.
// Unfortunately the notation has not yet been standardised, and implementations have signfiicant differences.
// This one aims to satisfy the existing consensus as represented by several test suites.
//
// For the detailed syntax, run
//
//	go doc jsonpath/syntax
//
// A given JSONpath expression (in textual form) is first transformed by Compile or MustCompile to
// a pointer to a JSONPath value that can then be applied repeatedly to JSON values using its Eval method.
// Compilation checks that the expression is valid JSONpath syntax as defined above.
package jsonpath

import (
	"strconv"

	"github.com/forsyth/jsonpath/mach"
	"github.com/forsyth/jsonpath/paths"
)

// JSONPath represents a compiled  JSONpath expression.
// It is safe for concurrent use by goroutines.
type JSONPath struct {
	expr string        // as passed to Compile
	path *paths.Path   // the parsed expression
	prog *mach.Program // the program for the abstract machine
}

// String returns the source text used to compile the JSONpath expression.
func (path *JSONPath) String() string {
	return path.expr
}

// Compile parses a JSONpath expression and, if it is syntactically valid, returns a JSONPath value
// that allows repeated evaluation of that expression against a given JSON value.
// If the expression is not valid JSONPath, Compile instead returns only an error.
func Compile(expr string) (*JSONPath, error) {
	path, err := paths.ParsePath(expr)
	if err != nil {
		return nil, err
	}
	prog, err := mach.Compile(path)
	if err != nil {
		return nil, err
	}
	return &JSONPath{expr: expr, prog: prog}, nil
}

// MustCompile is like Compile but panics if the expression is invalid.
func MustCompile(expr string) *JSONPath {
	p, err := Compile(expr)
	if err != nil {
		panic(`jsonpath: Compile(` + quote(expr) + `): ` + err.Error())
	}
	return p
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// Eval evaluates a previously-compiled JSONpath expression against a given JSON value
// (the root of a document), as returned by encoding/json.Decoder.Unmarshal.
// It returns a slice containing the list of JSON values selected by the path expression.
// If a run-time error occurs, for instance an invalid dynamic regular expression,
// Eval stops and returns only an error.
// Eval may be used concurrently.
//
// Path expressions contain boolean filter expressions of the form ?(expr), and other
// numeric, string or boolean expressions of the form (expr). The expression language
// is the same for each, containing a subset of JavaScript's expression and logical
// operators, and a match operator =~ (subject-string =~ /re/ or subject-string =~ regexp-string).
// The equality operators == and != apply JavaScript's equality rules.
// Boolean contexts apply JavaScript's falsy rules, and || and && also behave as in JavaScript.
// For instance (a || b) yields the value of b if a is falsy, not necessarily a Boolean value as in Go and other languages.
// JavaScript's implicit conversion rules are also applied when boolean, string and numeric values meet.
// Note that expressions also handle failure (eg, failing to find a key in an object, or selecting from a non-object)
// by propagating null values, which can then be detected and handled by using || and && as one might in JavaScript.
func (path *JSONPath) Eval(root interface{}) ([]interface{}, error) {
	return path.prog.Run(root)
}
