BINARY := gdt-assets
GO := go

.PHONY: build install clean

build:
	$(GO) build -o $(BINARY) .

install: build
	install -m 755 $(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || \
	install -m 755 $(BINARY) $(HOME)/go/bin/$(BINARY)

clean:
	rm -f $(BINARY)
