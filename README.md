# nosleep

A tiny macOS shell script for toggling system sleep with `pmset`.

## Usage

```bash
chmod +x nosleep.sh
./nosleep.sh status
./nosleep.sh on
./nosleep.sh off
./nosleep.sh help
```

## What It Does

- `on` runs `sudo pmset -a disablesleep 1`
- `off` runs `sudo pmset -a disablesleep 0`
- `status` reads the current `disablesleep` value from `pmset`
- `help` shows the usage message

## Notes

- macOS only
- `on` and `off` use `sudo`, so they may prompt for an administrator password
- running the script with no arguments shows the usage message
- `-a` applies the setting to all power profiles

## Optional Install

If you want to call it as `nosleep` from anywhere:

```bash
chmod +x nosleep.sh
ln -s "$(pwd)/nosleep.sh" /usr/local/bin/nosleep
```

The install step usually does not require a password, but creating the symlink may need one if `/usr/local/bin` is not writable by your user.

## Skip Password Prompts

To let specific users run `on` / `off` without ever being prompted for a password, add a `sudoers` rule:

```bash
# Allow all admin-group users to run pmset without password
echo "%admin ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep
```

Or limit it to a single user:

```bash
echo "yourusername ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep
```

This creates a drop-in file under `/etc/sudoers.d/`, which is safer than editing `/etc/sudoers` directly. The `status` command does not use `sudo`, so it is unaffected.
