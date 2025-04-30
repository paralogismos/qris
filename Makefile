# Simple Go Makefile
.POSIX:

P =                               # name of package to build provided by user
MAINDIR = cmd/$(P)                # assume cmd/<package>/main.go structure
MAIN = cmd/$(P)/main.go
EXECNAME = $(P).exe               # adds .exe on Windows
IPATH = ${GOPATH}\bin
IEXEC = $(GOPATH)\bin\$(EXECNAME)

.PHONY: fmt vet build install
build: fmt vet
	go build -o $(EXECNAME) $(MAIN)

vet: fmt
	go vet ./...

fmt:
	go fmt ./...

install: fmt vet
	go install -C $(MAINDIR)
	@printf '%s installed in %s\n' $(EXECNAME) $(IPATH)

.PHONY: clean uninstall
clean:
	rm -f $(EXECNAME)

uninstall:
	rm -f $(IEXEC)
	@printf '%s removed from %s\n' $(EXECNAME) $(IPATH)
