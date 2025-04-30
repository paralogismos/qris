# Simple Go Makefile
.POSIX:

P =                               # name of package to build provided by user
MAINDIR = cmd/$(P)                # assume cmd/<package>/main.go structure
EXECNAME = $(P).exe               # adds .exe on Windows
IPATH = ${GOPATH}\bin

.PHONY: fmt vet build install
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: fmt vet
	@pushd $(MAINDIR) && \
	go build         && \
	popd

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
	@pushd $(MAINDIR) && \
	go clean         && \
	popd

uninstall:
	@pushd $(MAINDIR) && \
	go clean -i      && \
	popd
