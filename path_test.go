package JSONPath

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

const testFile = "testdata/t1"

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
		fmt.Printf("%s -> ", sam)
		path, err := ParsePath(sam)
		if err != nil {
			fmt.Printf("!%s\n", err)
			continue
		}
		for _, el := range path {
			fmt.Printf(" %s", el)
		}
		fmt.Printf("\n")
	}
	if err := samples.Err(); err != nil {
		t.Fatalf("error reading %s (line %d): %s", testFile, lno, err)
	}
}

func (s *Step) String() string {
	doc := s.op.GoString()
	if len(s.args) > 0 {
		doc += "("
		for i, a := range s.args {
			if i > 0 {
				doc += ","
			}
			doc += fmt.Sprintf("%#v", a)
		}
		doc += ")"
	}
	return doc
}
