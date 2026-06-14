# nosleep

A tiny macOS shell script for toggling system sleep with `pmset`.

## Usage

```bash
chmod +x nosleep.sh
./nosleep.sh status
./nosleep.sh on
./nosleep.sh off
./nosleep.sh help
./nosleep.sh setup
```

## What It Does

- `on` runs `sudo pmset -a disablesleep 1`
- `off` runs `sudo pmset -a disablesleep 0`
- `status` reads the current `disablesleep` value from `pmset`
- `help` shows the usage message
- `setup` installs a sudoers rule so `on`/`off` stop asking for a password

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

The easiest way:

```bash
./nosleep.sh setup
```

This creates a sudoers drop-in file for the current user. You'll enter your password once during setup — after that, `on` and `off` work without prompting.

You can also do it manually for a single user or an entire group:

```bash
# Allow all admin-group users to run pmset without password
echo "%admin ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep
```

Or limit it to a single user:

```bash
echo "yourusername ALL=(ALL) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1" | sudo tee /etc/sudoers.d/nosleep
```

This creates a drop-in file under `/etc/sudoers.d/`, which is safer than editing `/etc/sudoers` directly. The `status` command does not use `sudo`, so it is unaffected.
