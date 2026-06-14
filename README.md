# nosleep

A tiny macOS shell script for toggling system sleep with `pmset`.

It is intended for cases where you need a system-level sleep toggle, including lid-closed use, rather than a temporary per-process keep-awake command.

## Usage

```bash
chmod +x nosleep.sh
./nosleep.sh on
./nosleep.sh off
./nosleep.sh status
./nosleep.sh setup
./nosleep.sh help
```

## What It Does

- `on` runs `sudo pmset -a disablesleep 1`
- `off` runs `sudo pmset -a disablesleep 0`
- `status` reads the current `disablesleep` value from `pmset`
- `setup` installs a sudoers rule so `on`/`off` stop asking for a password
- `help` shows the usage message

## Example Scenarios

- Running a long local job or server while the Mac is closed and away from a charger.
- Allowing longer-running work such as app builds, model training, or other compute-heavy tasks to continue while you are in transit or moving between meetings.
- Keeping a remote session, file transfer, or other background task alive when normal sleep would interrupt it.
- Temporarily preventing sleep on a machine used as a small home server, lab machine, or automation host.
- Using a Mac in a setup where `caffeinate` is not sufficient and you intentionally need sleep disabled at the system level.

## How This Differs From `caffeinate`

`caffeinate` is commonly used to keep a Mac awake while a process is running or while the display remains open, but it does not change the underlying system sleep setting in the same way.

This script uses `pmset` to change the system `disablesleep` setting directly. In practice, that means it can continue preventing sleep in situations where `caffeinate` is not the right tool, including lid-closed use and operation on battery power without a charger connected.

## Safety

- This changes a system-level power setting, not just a single terminal session.
- If left enabled, your Mac can remain awake with the lid fully closed and while not connected to power.
- This can increase battery drain and may cause the laptop to become hot, especially if it is placed in a bag or another poorly ventilated space.
- Use it only when you intentionally need this behavior, and run `off` as soon as you are done.

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
