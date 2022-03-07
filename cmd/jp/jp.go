package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/forsyth/jsonpath"
)

var stdout *bufio.Writer // global only because os.Exit doesn't run defers

func main() {
	byLine := flag.Bool("l", false, "one JSON value per line, and result set on single line")
	//	useNumber := flag.Bool("n", false, "represent JSON numbers as integer, floating-point or string")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: jp [-l] pat [file ...]\n")
		os.Exit(2)
	}
	stdout = bufio.NewWriter(os.Stdout)
	defer stdout.Flush()
	jexp := flag.Arg(0)
	jpath, err := jsonpath.Compile(jexp)
	if err != nil {
		errorf("path %s: %s", quote(jexp), err.Error())
	}
	var reader func(*os.File, *jsonpath.JSONPath, *json.Encoder) error
	enc := json.NewEncoder(stdout)
	if *byLine {
		enc.SetIndent("", "") // one-line output
		reader = readLines
	} else {
		enc.SetIndent("", "\t")
		reader = readValues
	}
	if flag.NArg() > 1 {
		for _, file := range flag.Args()[1:] {
			fd, err := os.Open(file)
			if err != nil {
				errorf("%s: cannot open: %s", file, err.Error())
			}
			err = reader(fd, jpath, enc)
			fd.Close()
			if err != nil {
				errorf("%s:%s", file, err.Error())
			}
		}
	} else {
		err = reader(os.Stdin, jpath, enc)
		if err != nil {
			errorf("%s", err.Error())
		}
	}
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// readValues runs the JSONPath machine against a sequence of JSON values, across newlines, producing results as formatted JSON.
func readValues(fd *os.File, jpath *jsonpath.JSONPath, enc *json.Encoder) error {
	dec := json.NewDecoder(fd)
	// dec.UseNumber()
	for {
		var root interface{}
		off := dec.InputOffset()
		err := dec.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("#%d: decoding JSON: %w", off, err)
		}
		results, err := jpath.Eval(root)
		if err != nil {
			return fmt.Errorf("#%d: evaluation error: %w", off, err)
		}
		err = enc.Encode(results)
		if err != nil {
			return fmt.Errorf("#%d: encoding results: %w", off, err)
		}
	}
	return nil
}

// readLines runs the JSONPath machine against JSON values, one per line, also producing the results on a single line (thus 1:1).
func readLines(fd *os.File, jpath *jsonpath.JSONPath, enc *json.Encoder) error {
	input := bufio.NewScanner(fd)
	input.Split(bufio.ScanLines)
	for lno := 1; input.Scan(); lno++ {
		var root interface{}
		err := json.Unmarshal(input.Bytes(), &root)
		if err != nil {
			return fmt.Errorf("%d: decoding JSON: %w", lno, err)
		}
		results, err := jpath.Eval(root)
		if err != nil {
			return fmt.Errorf("%d: evaluation error: %w", lno, err)
		}
		err = enc.Encode(results)
		if err != nil {
			return fmt.Errorf("%d: encoding results: %w", lno, err)
		}
	}
	if err := input.Err(); err != nil {
		return fmt.Errorf("read error: %w", err)
	}
	return nil
}

func errorf(format string, args ...interface{}) {
	stdout.Flush()
	fmt.Fprintf(os.Stderr, "%s: ", os.Args[0])
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
