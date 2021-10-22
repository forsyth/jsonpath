package JSONPath

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// standard "book" example as initial test case

const testJSON = "testdata/book.json"

var testQueries []string = []string{
	"$.store",
	"$.*",
	"$.*.*",
	"$.store.bicycle",
	"$.store.book",
	"$.store.book[1]",
	"$.store.book[1].author",
	"$.store.book[1].price",
	"$.store.book[(1*3-2)].price",
	"$.store.book[((1*3-2)*0)].price",
	"$.store.book[((1*3-2)/0)].price",
	"$.store.book[([[5,0,4],2,3][0][1])].price",
	"$.store.book[([[5,0,1],2,3][0][2])].price",
	"$.store.book[([[5,0,3*4-3*4+1],2,3][0/0][2])].price",
	//"$.store.book[([[5,@.price-@.price,3*4-3*4+1],2,3][0/0][2])].price",
	//	"$.store.book[?(@.price > 8.95)]",
	"$.store.book[:]",
	"$.store.book[1:]",
	"$.store.book[-1:]",
}

// TestRun applies Program.Run to the standard "book" example with various paths.
func TestRun(t *testing.T) {
	js := loadJSON(testJSON, t)
	for i, q := range testQueries {
		path, err := ParsePath(q)
		if err != nil {
			t.Errorf("sample %d: %s: parse: %s", i, q, err)
			continue
		}
		prog, err := CompilePath(path)
		if err != nil {
			t.Errorf("sample %d: %s: compile: %s", i, q, err)
			continue
		}
		if testing.Verbose() {
			fmt.Printf("%s -> %s\n", q, progString(prog))
		}
		vals, err := prog.Run(js)
		if err != nil {
			t.Errorf("sample %d: %s: run: %s", i, q, err)
			continue
		}
		fmt.Printf("sample %d: %s: success\n", i, q)
		_ = vals
	}
}

// TestWalker runs the value walker on the JSON in the test file.
// TO DO: provide a reference value (file).
func TestWalker(t *testing.T) {
	js := loadJSON(testJSON, t)
	values := make(chan JSON)
	go walker(values, []JSON{js})
	for item := range values {
		if testing.Verbose() {
			fmt.Printf("%#v\n", item)
		}
	}
}

// run the engine against an external test suite of sorts.

const testSuiteFile = "testdata/test_suite.yaml"

// test case excluded by this implementation, and why.
type exclusion struct {
	id  string
	err string
}

var exclusions map[string]string = map[string]string{ // samples excluded by this implementation
	"dot_notation_with_key_root_literal": "unexpected $ at offset 2",                    // reject
	"filter_expression_with_subfilter":   "unexpected character '?' at offset 8",        // TO DO: consider nested filters
	"union_with_filter":                  "?(filter) cannot be in a union element list", // reject
	"union_with_wildcard_and_number":     "* cannot be in a union element list",         // reject
}

var differences map[string]string = map[string]string{ // samples where this implementation gives a known different result
	"bracket_notation_with_number_on_object": "[\"value\"]",
	"dot_notation_with_wildcard_on_object":   "[42,{\"key\":\"value\"},[0,1],\"string\"]",
	"dot_notation_with_number":               "[\"third\"]",
	"filter_expression_with_subtraction":     "[{\"key\":-50}]", // they take - inside the expression as part of the key
}

// TestSuite is the root of a YAML-encoded test suite in an existing format.
type TestSuite struct {
	Queries []Query `yaml:"queries"`
}

// Query is one of the test queries, with a description (ID), query string (Selector), source Document and result Consensus.
type Query struct {
	ID        string      `yaml:"id"`
	Selector  string      `yaml:"selector"`
	Document  interface{} `yaml:"document"`
	Consensus interface{} `yaml:"consensus"`
}

func (q *Query) noConsensus() bool {
	if s, ok := q.Consensus.(string); ok {
		return s == "NOT_SUPPORTED"
	}
	return false
}

func (q *Query) excluded() string {
	return exclusions[q.ID]
}

func (q *Query) runDifferent() string {
	return differences[q.ID]
}

// do a and b differ only in order, including duplicates?
func isReordered(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
Search:
	for _, va := range a {
		for i, vb := range b {
			if !used[i] && reflect.DeepEqual(va, vb) {
				used[i] = true
				continue Search
			}
		}
		return false
	}
	for i := range b {
		if !used[i] {
			return false
		}
	}
	return true
}

// TestTestSuite runs Program.Run over the test suite from [TO DO: URL].
func TestTestSuite(t *testing.T) {
	ts := loadYAML(testSuiteFile, t)
	for qno, query := range ts.Queries {
		if testing.Verbose() {
			fmt.Printf("%d: %s %s %s\n", qno, query.ID, query.Selector, jsonString(query.Document))
		}
		path, err := ParsePath(query.Selector)
		if err != nil {
			expected := query.excluded()
			if expected != err.Error() {
				t.Errorf("%s: sample %d: %s: parse %q: %s", testSuiteFile, qno, query.ID, query.Selector, err)
			}
			if testing.Verbose() {
				fmt.Printf("%d: %s: selector %q: got error (%s) expected: %s\n", qno, query.ID, query.Selector, err, expected)
			}
			continue
		}
		if query.noConsensus() {
			// skip ones for which no consensus
			continue
		}
		prog, err := CompilePath(path)
		if err != nil {
			t.Errorf("%s: sample %d: %s: compile %q: %s", testSuiteFile, qno, query.ID, query.Selector, err)
			continue
		}
		fmt.Printf("prog: %s\n", progString(prog))
		results, err := prog.Run(query.Document)
		if err != nil {
			t.Errorf("%s: sample %d: %s: run error: %s", testSuiteFile, qno, query.ID, err)
			continue
		}
		s2 := jsonString(results)
		note := ""
		if query.Consensus != nil {
			s1 := jsonString(query.Consensus)
			fmt.Printf("results0: %s\n", s1)
			if s1 != s2 {
				note = " (differ)"
				s3 := query.runDifferent()
				if s3 != "" {
					if s2 != s3 {
						t.Errorf("%s: sample %d: %s: got result (%s) expected (%s)", testSuiteFile, qno, query.ID, s2, s3)
					} else {
						note = "(differ, but ok)"
					}
				} else {
					a, ok1 := results.([]interface{})
					b, ok2 := query.Consensus.([]interface{})
					if ok1 && ok2 && isReordered(a, b) {
						note = "(differ, but reordered)"
					}
				}
			}
		}
		fmt.Printf("results1: %s %s\n", s2, note)
	}
}

// loadJSON returns the JSON in the given file or gives a fatal error.
func loadJSON(file string, t *testing.T) JSON {
	data := loadFile(file, t)
	var js interface{}
	err := json.Unmarshal(data, &js)
	if err != nil {
		t.Fatalf("%s: erroneous JSON: %s", file, err)
	}
	return js
}

// loadYAML returns the external TestSuite, which is in YAML format, for fun.
func loadYAML(file string, t *testing.T) *TestSuite {
	data := loadFile(file, t)
	var ts *TestSuite
	err := yaml.Unmarshal(data, &ts)
	if err != nil {
		t.Fatalf("%s: erroneous test suite: %s", file, err)
	}
	return ts
}

// loadFile loads the entire contents of a small test file, or gives a fatal error.
func loadFile(file string, t *testing.T) []byte {
	fd, err := os.Open(file)
	if err != nil {
		t.Fatalf("%s: cannot open: %s", file, err)
	}
	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatalf("%s: read error: %s", file, err)
	}
	return data
}
