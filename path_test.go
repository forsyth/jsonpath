package JSONPath

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

const testFile = "testdata/t1"
const separator = " -> " // separates input from desired output on same line

func TestPathParse(t *testing.T) {
	tfd, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("cannot open %s: %s", testFile, err)
	}
	samples := bufio.NewScanner(tfd)
	lno := 0
	for samples.Scan() {
		lno++
		sam := samples.Text()
		if sam == "" || sam[0] == '#' {
			// comment
			continue
		}
		// input -> desired-output
		sep := strings.Index(sam, separator)
		var proto string
		if sep > 0 { // if it's zero, assume that's input
			proto = sam[sep+len(separator):]
			sam = sam[0:sep]
		}
		fmt.Printf("%s -> ", sam)
		path, err := ParsePath(sam)
		if err != nil {
			fmt.Printf("!%s\n", err)
			if proto != "" {
				if proto[0] != '!' {
					t.Errorf("line %d, sample %s, got error %q, expected %s", lno, sam, err, proto)
				} else if err.Error() != proto {
					t.Errorf("line %d, sample %s, got error %q, expected error %q", lno, sam, err, proto)
				}
			}
			continue
		}
		for i, el := range path {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Printf("%s", el)
		}
		fmt.Printf("\n")
	}
	if err := samples.Err(); err != nil {
		t.Fatalf("error reading %s (line %d): %s", testFile, lno, err)
	}
}
