PREFIX ?= $(shell brew --prefix 2>/dev/null || echo /usr/local)

.PHONY: build test lint install install-man
build:
	go build -o orch ./cmd/orch

test:
	go vet ./...
	go test -race ./...

lint:
	golangci-lint run

install:
	go install ./cmd/orch

# Puts man/orch.1 where `man orch` finds it (Homebrew's man dir is on the default MANPATH).
install-man:
	install -d $(PREFIX)/share/man/man1
	install -m 644 man/orch.1 $(PREFIX)/share/man/man1/orch.1
