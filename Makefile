PREFIX ?= /usr/local

.PHONY: build test man clean install uninstall

build: s3sum

s3sum: $(wildcard *.go cmd/*.go) go.mod go.sum
	go build -o $@

test:
	go test ./...

man: man/s3sum.1

man/s3sum.1: $(wildcard *.go cmd/*.go doc/*.go) go.mod go.sum
	go run ./doc

clean:
	rm -f s3sum
	rm -rf man

install: s3sum man/s3sum.1
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 s3sum $(DESTDIR)$(PREFIX)/bin/s3sum
	install -d $(DESTDIR)$(PREFIX)/share/man/man1
	install -m 644 man/s3sum.1 $(DESTDIR)$(PREFIX)/share/man/man1/s3sum.1

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/s3sum
	rm -f $(DESTDIR)$(PREFIX)/share/man/man1/s3sum.1
