#!/usr/bin/env bash

set -euo pipefail

usage() {
  local script_name
  script_name="$(basename "$0")"

  cat <<EOF
Usage: ${script_name} [on|off|status|setup|help]

Commands:
  on      Disable system sleep on macOS
  off     Re-enable system sleep on macOS
  status  Show whether sleep is currently disabled
  setup   Install a sudoers rule so on/off stop asking for a password
  help    Show this message

Options:
  --json  Output machine-readable JSON (applies to on, off, status, setup)

Examples:
  ./${script_name} status
  ./${script_name} on
  ./${script_name} off
  ./${script_name} setup
  ./${script_name} status --json
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

# Check if --json flag is present in arguments
use_json() {
  local arg
  for arg in "$@"; do
    if [[ "$arg" == "--json" ]]; then
      return 0
    fi
  done
  return 1
}

main() {
  require_macos

  local script_name
  script_name="$(basename "$0")"

  # Extract command (first non-flag argument)
  local command="help"
  local arg
  for arg in "$@"; do
    if [[ "$arg" != "--" && "$arg" != --* ]]; then
      command="$arg"
      break
    fi
  done

  local json_output=false
  if use_json "$@"; then
    json_output=true
  fi

  case "$command" in
    on)
      if $json_output; then
        sudo pmset -a disablesleep 1 && echo '{"ok":true,"action":"on"}'
      else
        echo "Disabling sleep..."
        sudo pmset -a disablesleep 1
      fi
      ;;
    off)
      if $json_output; then
        sudo pmset -a disablesleep 0 && echo '{"ok":true,"action":"off"}'
      else
        echo "Enabling sleep..."
        sudo pmset -a disablesleep 0
      fi
      ;;
    status)
      local disabled
      disabled="$(sleep_disabled_value)"

      if $json_output; then
        if [[ "${disabled:-0}" == "1" ]]; then
          echo '{"state":"awake","disablesleep":1}'
        else
          echo '{"state":"normal","disablesleep":0}'
        fi
      else
        if [[ "${disabled:-0}" == "1" ]]; then
          echo "Sleep is DISABLED (nosleep is ON)."
        else
          echo "Sleep is ENABLED (nosleep is OFF)."
        fi
      fi
      ;;
    setup)
      if $json_output; then
        echo "${USER} ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep > /dev/null && echo "{\"ok\":true,\"action\":\"setup\",\"user\":\"${USER}\"}"
      else
        echo "Setting up password-less nosleep for user '${USER}'..."
        echo "${USER} ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep > /dev/null
        echo "Done. You can now run '${script_name} on' and '${script_name} off' without a password."
      fi
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      echo "Error: invalid argument '${command}'." >&2
      usage >&2
      exit 1
      ;;
  esac
}

main "$@"
