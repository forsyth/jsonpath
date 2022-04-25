SHELL=/bin/rc
STRINGS=\
	paths/op_string.go \

all:V: $STRINGS
	go build
	go build ./cmd/jpath
	go vet . ./paths ./mach ./cmd/jpath

paths/op_string.go:D: paths/ops.go
	go generate paths/ops.go

fmt:V:
	go fmt . ./paths ./mach ./cmd/jpath

test:V:
	go test . ./paths ./mach
