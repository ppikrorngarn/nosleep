#!/usr/bin/env bash

set -euo pipefail

usage() {
  local script_name
  script_name="$(basename "$0")"

  cat <<EOF
Usage: ${script_name} [on|off|status|help]

Commands:
  on      Disable system sleep on macOS
  off     Re-enable system sleep on macOS
  status  Show whether sleep is currently disabled
  help    Show this message

Examples:
  ./${script_name} status
  ./${script_name} on
  ./${script_name} off
EOF
}

require_macos() {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "Error: nosleep.sh only supports macOS." >&2
    exit 1
  fi
}

sleep_disabled_value() {
  pmset -g | awk '$1 == "SleepDisabled" { print $2; exit }'
}

main() {
  require_macos

  local script_name
  script_name="$(basename "$0")"

  case "${1:-help}" in
    on)
      echo "Disabling sleep..."
      sudo pmset -a disablesleep 1
      ;;
    off)
      echo "Enabling sleep..."
      sudo pmset -a disablesleep 0
      ;;
    status)
      local disabled
      disabled="$(sleep_disabled_value)"

      if [[ "${disabled:-0}" == "1" ]]; then
        echo "Sleep is DISABLED (nosleep is ON)."
      else
        echo "Sleep is ENABLED (nosleep is OFF)."
      fi
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      echo "Error: invalid argument '${1}'." >&2
      usage >&2
      exit 1
      ;;
  esac
}

main "${1:-help}"
