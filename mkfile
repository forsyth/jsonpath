SHELL=/bin/rc
STRINGS=\
	paths/op_string.go \

all:V: $STRINGS
	go build
	go vet . ./paths ./mach

paths/op_string.go:D: paths/ops.go
	go generate paths/ops.go

fmt:V:
	go fmt . ./paths ./mach

test:V:
	go test . ./paths ./mach
