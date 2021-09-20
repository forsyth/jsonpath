package JSONPath

import (
	"fmt"
	"testing"
)

type el struct {
	op  Op
	val interface{}
}

type lexOutput struct {
	s   string
	ops []el
}

var samples []lexOutput = []lexOutput{
	lexOutput{"$", []el{{op: Oroot}, el{op: Oeof}}},
	lexOutput{"$.store.book[(@.length-1)].title",
		[]el{el{op: Oroot}, el{op: Odot}, el{op: Oid, val: "store"}, el{op: Odot}, el{op: Oid, val: "book"}, el{op: Obra}, el{op: Olpar},
		      el{op: Ocurrent}, el{op: Odot}, el{op: Oid, val: "length"}, el{op: Osub}, el{op: Oint, val: 1}, el{op: Orpar}, el{op: Oket}, el{op: Odot}, el{op: Oid}, el{op: Oeof},
		},
	},
	lexOutput{"$.store.book[?(@.price < 10)].title",
		[]el{el{op: Oroot}, el{op: Odot}, el{op: Oid}, el{op: Odot}, el{op: Oid}, el{op: Obra}, el{op: Ofilter},
		      el{op: Ocurrent}, el{op: Odot}, el{op: Oid}, el{op: Olt}, el{op: Oint, val: 10}, el{op: Orpar}, el{op: Oket}, el{op: Odot}, el{op: Oid, val: "title"}, el{op: Oeof},
		},
	},
	lexOutput{"$.['store'].book[?(@.price < 10)].title",
		[]el{el{op: Oroot}, el{op: Odot}, el{op: Obra}, el{op: Ostring, val: "store"}, el{op: Oket},  el{op: Odot}, el{op: Oid, val: "book"}, el{op: Obra}, el{op: Ofilter},
		      el{op: Ocurrent}, el{op: Odot}, el{op: Oid, val: "price"}, el{op: Olt}, el{op: Oint, val: 10}, el{op: Orpar}, el{op: Oket}, el{op: Odot}, el{op: Oid, val: "title"}, el{op: Oeof},
		},
	},
	lexOutput{"$..book[(@.length-1)]",
		[]el{el{op: Oroot},  el{op: Onest},  el{op: Oid, val:"book"}, el{op: Obra},  el{op: Olpar},  el{op: Ocurrent},  el{op: Odot},  el{op: Oid, val: "length"}, el{op: Osub},  el{op: Oint, val: 1},  el{op: Orpar},  el{op: Oket},  el{op: Oeof}, 
		},
	},
}

func TestLex(t *testing.T) {
	for i, sam := range samples {
		rdr := &rd{sam.s, 0}
		fmt.Printf("%s -> ", sam.s)
		for j, el := range sam.ops {
			op, val, err := lex(rdr)
			print(op, val, err)
			if op != el.op || op != Oerror && err != nil {
				t.Errorf("sample %d el %d, got %v (%#v %v) expected %v (%#v)", i, j, op, val, err, el.op, el.val)
				break
			}
		}
		fmt.Printf("\n")
		if rdr.look() != eof {
			t.Errorf("sample %d, not reached Oeof", i)
			for {
				op, val, err := lex(rdr)
				print(op, val, err)
				if op == Oeof || op == Oerror {
					break
				}
			}
			fmt.Printf("\n")
		}
	}
}

func print(op Op, val interface{}, err error) {
	fmt.Printf(" %v", op)
	if op.hasVal() {
		fmt.Printf("[%#v]", val)
	}
	if err != nil {
		fmt.Printf("!%s", err)
	}
}
