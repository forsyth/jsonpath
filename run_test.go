package JSONPath

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// standard "book" example as initial test case

const testJSON = "testdata/book.json"

var testQueries = []string{
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
		prog, err := path.Compile()
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

var exclusions = map[string]string{ // samples excluded by this implementation, usually unacceptable syntax
	"dot_notation_with_key_root_literal":                                  "unexpected $ at offset 2",                    // reject
	"filter_expression_with_subfilter":                                    "unexpected character '?' at offset 8",        // TO DO: consider nested filters
	"union_with_filter":                                                   "?(filter) cannot be in a union element list", // reject
	"union_with_wildcard_and_number":                                      "* cannot be in a union element list",         // reject
	"bracket_notation_with_empty_path":                                    "unexpected ] at offset 2",
	"bracket_notation_with_quoted_string_and_unescaped_single_quote":      "expected \"]\" at offset 14, got identifier",
	"bracket_notation_with_two_literals_separated_by_dot":                 "expected \"]\" at offset 7, got .",
	"bracket_notation_with_two_literals_separated_by_dot_without_quotes":  "expected \"]\" at offset 5, got .",
	"dot_bracket_notation":                                                "unexpected [ at offset 2",
	"dot_bracket_notation_with_double_quotes":                             "unexpected [ at offset 2",
	"dot_bracket_notation_without_quotes":                                 "unexpected [ at offset 2",
	"dot_notation_with_double_quotes":                                     "unexpected string literal at offset 6",
	"dot_notation_with_double_quotes_after_recursive_descent":             "unexpected string literal at offset 7",
	"dot_notation_with_empty_path":                                        "unexpected end of expression at offset 1",
	"dot_notation_with_single_quotes":                                     "unexpected string literal at offset 6",
	"dot_notation_with_single_quotes_after_recursive_descent":             "unexpected string literal at offset 7",
	"dot_notation_with_single_quotes_and_dot":                             "unexpected string literal at offset 11",
	"dot_notation_without_root":                                           "expected \"$\" at offset 2, got identifier",
	"filter_expression_with_empty_expression":                             "unexpected token ) in expression term",
	"filter_expression_with_equals_array_for_array_slice_with_range_1":    "unexpected character ':' at offset 7",
	"filter_expression_with_equals_array_for_dot_notation_with_star":      "expected identifier in '.' selection",
	"filter_expression_with_equals_number_for_array_slice_with_range_1":   "unexpected character ':' at offset 7",
	"filter_expression_with_equals_number_for_bracket_notation_with_star": "unexpected token * in expression term",
	"filter_expression_with_equals_number_for_dot_notation_with_star":     "expected identifier in '.' selection",
	"filter_expression_with_equals_object":                                "unexpected character '{' at offset 9",
	"filter_expression_with_single_equal":                                 "expected \")\" at offset 9, got =",
	"filter_expression_with_triple_equal":                                 "unexpected token = in expression term",
	"filter_expression_without_parens":                                    "unexpected char '@' after '(' at offset 2",
	"parens_notation":                                                     "unexpected token (",
	"recursive_descent":                                                   "unexpected end of expression at offset 2",
	"recursive_descent_after_dot_notation":                                "unexpected end of expression at offset 6",
}

var differences = map[string]string{ // samples where this implementation gives a known different result
	"bracket_notation_with_number_on_object": "[\"value\"]",
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
	for _, v := range used {
		if !v {
			// some element of b wasn't matched
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
			// skip ones with no consensus
			continue
		}
		prog, err := path.Compile()
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
		note := " "
		if query.Consensus != nil {
			s1 := jsonString(query.Consensus)
			fmt.Printf("results0: %s\n", s1)
			if s1 != s2 {
				b, ok := query.Consensus.([]interface{})	// it's always a set
				if ok && isReordered(results, b) {
					note = " (differ, but reordered)"
				} else {
					note = " (differ)"
					s3 := query.runDifferent()
					if s3 != "" {
						if s2 != s3 {
							t.Errorf("%s: sample %d: %s: got result (%s) expected (%s)", testSuiteFile, qno, query.ID, s2, s3)
						} else {
							note = " (differ, but ok)"
						}
					}
				}
			}
		}
		fmt.Printf("results1: %s%s\n", s2, note)
	}
}

// second test suite format from Daniel A Parker

type PathTest struct {
	Given	JSON  `json:"given"`	// either [] or {}
	Cases	[]TestCase `json:"cases"`
}

type TestCase struct {
	Source	string	`json:"source"`	// eg, github...
	Comment	string	`json:"comment"`	// what it tests
	Error	interface{}	`json:"error"`	// correct response is an error: sometimes a string, sometimes a bool
	Skip	bool 	`json:"skip"`	// test marked to be skipped by this implementation
	Expression	string	`json:"expression"`	// path expression
	Nodups	bool	`json:"nodups"`	// remove duplicates from output set (not implemented)
	Result	[]JSON	`json:"result"`
	Path	[]string	`json:"path"`	// not used: expression for each subpath to result
}

// directory of tests in Parker's JSON format
const testParker = "testdata/group1"

// TestParkTests runs the set of tests adapted from Parker's C# implementation.
func TestParkerTests(t *testing.T) {
	dir, err := ioutil.ReadDir(testParker)
	if err != nil {
		t.Fatalf("%s: cannot read test directory: %s", testParker, err)
	}
	for _, file := range dir {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		fileName := testParker+"/"+file.Name()
		tests := loadParkerTest(fileName, t)
		fmt.Printf("%s: %d\n", fileName, len(tests))
		for tno, test := range tests {
			given := jsonString(test.Given)
			fmt.Printf("example %d: given: %s\n", tno, given)
			fmt.Printf("GIVEN: %#v\n", test.Given)
			for tcno, tc := range test.Cases {
				if tc.Comment != "" {
					fmt.Printf("comment: %s", tc.Comment)
					fmt.Printf("\n")
				}
				fmt.Printf("expr: %s", tc.Expression)
				if tc.Skip {
					fmt.Printf(" [skipped]\n")
					continue
				}
				fmt.Printf("\n")
				path, err := ParsePath(tc.Expression)
				if err != nil {
					if tc.Error == nil {
						t.Errorf("%s: test %d.%d: path %s: %s", fileName, tno, tcno, tc.Expression, err)
					} else {
						t.Logf("%s: test %d.%d: path %s: %s [expected]", fileName, tno, tcno, tc.Expression, err)
					}
					continue
				} else if tc.Error != nil {
					t.Errorf("%s: test %d.%d: path %s: error expected, but passed", fileName, tno, tcno, tc.Expression)
					continue
				}
				prog, err := path.Compile()
				if err != nil {
					t.Errorf("%s: path %s: compile: %s", fileName, tc.Expression, err)
				}
				results, err := prog.Run(test.Given)
				if err != nil {
					t.Errorf("%s: path %s: run %.40s...: %s", fileName, tc.Expression, string(given), err)
				}
				fmt.Printf("results: %d\n", len(results))
				if tc.Result != nil {
					fmt.Printf("results0: %s\n", jsonString(tc.Result))
					s2 := jsonString(results)
					fmt.Printf("results1: %s\n", s2);
					if !reflect.DeepEqual(results, tc.Result) && !isReordered(results, tc.Result) {
						t.Errorf("%s: path %s: run %.40s...: wanted %s, got %s", fileName, tc.Expression, string(given), jsonString(tc.Result), s2)
					}
				}
			}
		}
	}
}

// loadParkerTest returns the set of Parker-format tests in the given file or gives a fatal error.
func loadParkerTest(file string, t *testing.T) []PathTest {
	data := loadFile(file, t)
	var js []PathTest
	err := json.Unmarshal(data, &js)
	if err != nil {
		t.Fatalf("%s: erroneous Parker test: %s", file, err)
	}
	return js
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
