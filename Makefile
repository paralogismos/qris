# Simple Go Makefile
.POSIX:

P = qris                          # name of package to build
MAINDIR = cmd/$(P)                # assume cmd/<package>/main.go structure
EXECNAME = $(P).exe               # adds .exe on Windows
IPATH = ${GOPATH}\bin

.PHONY: fmt vet build install
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: fmt vet
	cd $(MAINDIR) && \
	go build         && \
	cd ../..

# Tried to display the installation location.
#
# I couldn't make any version of this approach work:
# go install -C -n $(MAINDIR) 2>&1 | grep 'mv $$WORK' | cut -d " " -f3
#
# This works, but may not always be accurate.
# @printf '%s installed in %s\n' $(EXECNAME) $(IPATH)

install: fmt vet
	go install -C $(MAINDIR)

.PHONY: clean uninstall
clean:
	@cd $(MAINDIR) && \
	go clean         && \
	cd ../..

uninstall:
	@cd $(MAINDIR) && \
	go clean -i      && \
	cd ../..
