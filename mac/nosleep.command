#!/bin/bash

# NoSleep TUI Launcher for macOS
# Double-click this file to run the NoSleep TUI application

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Change to the script directory
cd "$SCRIPT_DIR"

# Make sure the TUI binary is in the right place
TUI_BINARY="$SCRIPT_DIR/../cmd/nosleep-tui/nosleep-tui"

# Check if the TUI binary exists
if [[ ! -f "$TUI_BINARY" ]]; then
    echo "Error: TUI binary not found at $TUI_BINARY"
    echo "Please make sure the nosleep-tui binary is in the cmd/nosleep-tui directory."
    echo ""
    echo "Press Enter to exit..."
    read
    exit 1
fi

# Make the TUI binary executable (in case it wasn't)
chmod +x "$TUI_BINARY"

# Run the TUI application
exec "$TUI_BINARY"

# If the TUI exits with an error, show a message
if [[ $? -ne 0 ]]; then
    echo ""
    echo "The application exited with an error."
    echo "Press Enter to exit..."
    read
fi
