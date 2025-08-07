.PHONY: install
install:
	cargo install rust2go-cli

# Generate go bindings
.PHONY: generate
generate: install
	rust2go-cli --src src/lib.rs --dst go/gen.go

# Clippy
.PHONY: clippy
clippy: fmt
	cargo clippy --fix --allow-dirty

# fmt
.PHONY: fmt
fmt:
	cargo fmt