package jsonpath

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
	building := testing.Verbose() // building testdata file
	samples := bufio.NewScanner(tfd)
	lno := 0
	for samples.Scan() {
		lno++
		sam := samples.Text()
		if sam == "" || sam[0] == '#' {
			// comment
			if building {
				fmt.Printf("%s\n", sam)
			}
			continue
		}
		// input -> desired-output
		sep := strings.Index(sam, separator)
		var proto string
		if sep > 0 { // if it's zero, assume that's input
			proto = sam[sep+len(separator):]
			sam = sam[0:sep]
		}
		if building {
			fmt.Printf("%s -> ", sam)
		}
		path, err := ParsePath(sam)
		if err != nil {
			if building {
				fmt.Printf("!%s\n", err)
			}
			if proto != "" {
				if proto[0] != '!' {
					t.Errorf("line %d, sample %s, got error %q, expected %s", lno, sam, err, proto)
				} else if err.Error() != proto[1:] {
					t.Errorf("line %d, sample %s, got error %q, expected error %q", lno, sam, err, proto[1:])
				}
			}
			continue
		}
		code, err := codePath(path)
		if err != nil {
			t.Errorf("line %d, sample %s, compilation error: %s", lno, sam, err)
			continue
		}
		if building {
			fmt.Print(code)
			fmt.Printf("\n")
		}
		if code != proto {
			t.Errorf("line %d, coded results disagree, got %q expected %q", lno, code, proto)
		}
	}
	if err := samples.Err(); err != nil {
		t.Fatalf("error reading %s (line %d): %s", testFile, lno, err)
	}
}

// build a program for the Path and return a readable version as a string
func codePath(path Path) (string, error) {
	prog, err := path.Compile()
	if err != nil {
		return "", err
	}
	return progString(prog), nil
}

func progString(prog *Program) string {
	var sb strings.Builder
	for i, val := range prog.vals {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(val.String())
	}
	if len(prog.vals) > 0 {
		sb.WriteByte(' ')
	}
	for i, ord := range prog.orders {
		if i > 0 {
			sb.WriteByte(' ')
		}
		op := ord.op()
		sb.WriteString(trimOp(op))
		if op.IsLeaf() {
			if !op.HasVal() {
				continue
			}
			if ord.isSmallInt() {
				sb.WriteByte('(')
				sb.WriteString(fmt.Sprint(ord.smallInt()))
				sb.WriteByte(')')
			} else {
				sb.WriteByte('[')
				sb.WriteString(fmt.Sprint(ord.index()))
				sb.WriteByte(']')
			}
			continue
		}
		if ord.isSmallInt() && ord.smallInt() != 0 {
			sb.WriteByte('.')
			sb.WriteString(fmt.Sprint(ord.smallInt()))
		}
	}
	return sb.String()
}

func trimOp(op Op) string {
	name := opNames[op]
	if name == "" {
		panic("unknown op in opNames")
	}
	return name[2:]
}
