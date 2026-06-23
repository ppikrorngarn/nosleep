.PHONY: build clean run app

APP_NAME = NoSleep
APP_BUNDLE = $(APP_NAME).app
APP_DIR = build/$(APP_BUNDLE)

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
	@rm -rf build/
	@echo "==> Clean complete"

run: build
	@echo "==> Running nosleep-tui..."
	@./cmd/nosleep-tui/nosleep-tui

app: build
	@echo "==> Building $(APP_BUNDLE)..."
	@mkdir -p $(APP_DIR)/Contents/MacOS
	@mkdir -p $(APP_DIR)/Contents/Resources
	@cp mac/Info.plist $(APP_DIR)/Contents/Info.plist
	@printf 'APPL????' > $(APP_DIR)/Contents/PkgInfo
	@cp mac/launcher.sh $(APP_DIR)/Contents/MacOS/NoSleep
	@chmod +x $(APP_DIR)/Contents/MacOS/NoSleep
	@cp cmd/nosleep-tui/nosleep-tui $(APP_DIR)/Contents/Resources/nosleep-tui
	@chmod +x $(APP_DIR)/Contents/Resources/nosleep-tui
	@if [ -f mac/AppIcon.icns ]; then cp mac/AppIcon.icns $(APP_DIR)/Contents/Resources/AppIcon.icns; fi
	@codesign --force --deep --sign - $(APP_DIR)
	@echo "==> App bundle: $(APP_DIR)"
	@echo "==> Copy to /Applications:  cp -r $(APP_DIR) /Applications/"
