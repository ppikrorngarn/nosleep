.PHONY: build clean run

# Default target
all: build

build:
	@echo "==> Generating and building nosleep-tui..."
	@cd cmd/nosleep-tui && go generate ./...
	@cd cmd/nosleep-tui && go build -o nosleep-tui
	@echo "==> Build complete: cmd/nosleep-tui/nosleep-tui"

clean:
	@echo "==> Cleaning up..."
	@rm -f cmd/nosleep-tui/nosleep-tui
	@rm -f cmd/nosleep-tui/nosleep.sh
	@echo "==> Clean complete"

run: build
	@echo "==> Running nosleep-tui..."
	@./cmd/nosleep-tui/nosleep-tui
