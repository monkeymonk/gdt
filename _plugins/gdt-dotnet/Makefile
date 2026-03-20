BINARY := gdt-dotnet
INSTALL_DIR := $(HOME)/.gdt/plugins/gdt-dotnet

.PHONY: build install clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/
	cp plugin.toml $(INSTALL_DIR)/

clean:
	rm -f $(BINARY)
