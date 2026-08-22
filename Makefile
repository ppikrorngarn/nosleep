.PHONY: build clean run tui-app gui run-gui release

TUI_BIN = build/nosleep-tui
TUI_APP_BUNDLE = NoSleep-TUI.app
TUI_APP_DIR = build/$(TUI_APP_BUNDLE)
TUI_APP_EXEC = NoSleep-TUI
TUI_DIR = NoSleep-TUI

GUI_DIR = NoSleep-GUI
GUI_DERIVED = $(GUI_DIR)/DerivedData
GUI_APP = $(GUI_DERIVED)/Build/Products/Release/NoSleep.app
GUI_BUNDLE = build/NoSleep-GUI.app

# Default target
all: build

build:
	@echo "==> Generating and building nosleep-tui..."
	@mkdir -p build
	@cd cmd/nosleep-tui && go generate ./...
	@cd cmd/nosleep-tui && go build -o "$(CURDIR)/$(TUI_BIN)"
	@echo "==> Build complete: $(TUI_BIN)"

clean:
	@echo "==> Cleaning up..."
	@rm -f cmd/nosleep-tui/nosleep-tui
	@rm -f cmd/nosleep-tui/nosleep.sh
	@rm -rf build/
	@rm -rf $(GUI_DERIVED)
	@rm -rf $(GUI_DIR)/NoSleepGUI.xcodeproj
	@echo "==> Clean complete"

run: build
	@echo "==> Running nosleep-tui..."
	@./$(TUI_BIN)

tui-app: build
	@echo "==> Building $(TUI_APP_BUNDLE)..."
	@mkdir -p $(TUI_APP_DIR)/Contents/MacOS
	@mkdir -p $(TUI_APP_DIR)/Contents/Resources
	@cp $(TUI_DIR)/Info.plist $(TUI_APP_DIR)/Contents/Info.plist
	@printf 'APPL????' > $(TUI_APP_DIR)/Contents/PkgInfo
	@cp $(TUI_DIR)/launcher.sh $(TUI_APP_DIR)/Contents/MacOS/$(TUI_APP_EXEC)
	@chmod +x $(TUI_APP_DIR)/Contents/MacOS/$(TUI_APP_EXEC)
	@cp $(TUI_BIN) $(TUI_APP_DIR)/Contents/Resources/nosleep-tui
	@chmod +x $(TUI_APP_DIR)/Contents/Resources/nosleep-tui
	@if [ -f $(TUI_DIR)/AppIcon.icns ]; then cp $(TUI_DIR)/AppIcon.icns $(TUI_APP_DIR)/Contents/Resources/AppIcon.icns; fi
	@codesign --force --deep --sign - $(TUI_APP_DIR)
	@echo "==> App bundle: $(TUI_APP_DIR)"
	@echo "==> Copy to /Applications:  cp -r $(TUI_APP_DIR) /Applications/"

gui:
	@echo "==> Generating NoSleep GUI Xcode project..."
	@command -v xcodegen >/dev/null 2>&1 || { echo "Error: xcodegen is required. Install with: brew install xcodegen"; exit 1; }
	@cd $(GUI_DIR) && xcodegen generate
	@echo "==> Building NoSleep GUI..."
	@xcodebuild -scheme NoSleepGUI -configuration Release -project $(GUI_DIR)/NoSleepGUI.xcodeproj -derivedDataPath $(GUI_DERIVED) -destination 'platform=macOS'
	@rm -rf $(GUI_BUNDLE)
	@mkdir -p build
	@ditto "$(GUI_APP)" "$(GUI_BUNDLE)"
	@codesign --force --deep --sign - --entitlements $(GUI_DIR)/NoSleep-GUI.entitlements "$(GUI_BUNDLE)"
	@echo "==> App bundle: $(GUI_BUNDLE)"
	@echo "==> Copy to /Applications:  cp -r $(GUI_BUNDLE) /Applications/"

run-gui: gui
	@echo "==> Opening NoSleep GUI..."
	@open "$(GUI_BUNDLE)"

release: clean tui-app gui
	@echo "==> Release build complete"
	@echo "==> TUI binary: $(TUI_BIN)"
	@echo "==> TUI app:    $(TUI_APP_DIR)"
	@echo "==> GUI app:    $(GUI_BUNDLE)"
