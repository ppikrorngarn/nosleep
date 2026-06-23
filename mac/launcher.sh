#!/bin/bash

# NoSleep.app launcher
# Opens Terminal.app and runs the bundled nosleep-tui binary

RESOURCES_DIR="$(cd "$(dirname "$0")/../Resources" && pwd)"
TUI_BINARY="$RESOURCES_DIR/nosleep-tui"

if [[ ! -f "$TUI_BINARY" ]]; then
    osascript -e 'display dialog "NoSleep TUI binary not found.\nPlease reinstall NoSleep." buttons {"OK"} default button "OK" with icon stop with title "NoSleep"'
    exit 1
fi

chmod +x "$TUI_BINARY"

# Open Terminal.app and run the TUI in a new window
osascript <<EOF
tell application "Terminal"
    activate
    do script "$TUI_BINARY"
end tell
EOF
