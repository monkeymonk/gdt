BINARY := gdt-ai

.PHONY: build install clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
