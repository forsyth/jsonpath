package jsonpath_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/forsyth/jsonpath"
)

var paths = []string{
	"$",
	"$.books",
	"$.books[?(@.author=='Adam Smith')].title",
	"$.books[0].author",
	"$.books[*].title",
}

var docs = []string{
	`{"books": [
		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	]}`,
}

func Example() {
	for _, s := range paths {

		// ParsePath parses a JSON path expression into a Path: a sequence of Steps
		jpath, err := jsonpath.Compile(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "example: path %q: %s\n", s, err)
			continue
		}
		fmt.Printf("pattern: %s\n", s)
		for _, doc := range docs {
			var d interface{}
			err = json.Unmarshal([]byte(doc), &d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "example: subject %q: %s\n", doc, err)
				continue
			}

			// jpath.Eval evaluates the path s on a given JSON value, the root of a document, returning
			// the list of resulting JSON values. The same JSONPath value can be reused, even concurrently.
			vals, err := jpath.Eval(d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "example: path %q: subject %q: %s\n", s, doc, err)
				continue
			}
			fmt.Printf("subject: %s ->\n", doc)
			results, err := json.Marshal(vals)
			if err != nil {
				fmt.Printf("!%s\n", err)
				continue
			}
			fmt.Println("\t", string(results))
		}
	}
	// Output:
	// pattern: $
	// subject: {"books": [
	// 		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
	// 		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	// 	]} ->
	// 	 [{"books":[{"author":"Evelyn Waugh","date":1928,"title":"Decline and Fall"},{"author":"Adam Smith","date":1776,"title":"Wealth of Nations"}]}]
	// pattern: $.books
	// subject: {"books": [
	// 		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
	// 		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	// 	]} ->
	// 	 [[{"author":"Evelyn Waugh","date":1928,"title":"Decline and Fall"},{"author":"Adam Smith","date":1776,"title":"Wealth of Nations"}]]
	// pattern: $.books[?(@.author=='Adam Smith')].title
	// subject: {"books": [
	// 		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
	// 		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	// 	]} ->
	// 	 ["Wealth of Nations"]
	// pattern: $.books[0].author
	// subject: {"books": [
	// 		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
	// 		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	// 	]} ->
	// 	 ["Evelyn Waugh"]
	// pattern: $.books[*].title
	// subject: {"books": [
	// 		{"title": "Decline and Fall", "author": "Evelyn Waugh", "date": 1928},
	// 		{"title": "Wealth of Nations", "author": "Adam Smith", "date": 1776}
	// 	]} ->
	// 	 ["Decline and Fall","Wealth of Nations"]
}
