# gwatch

A lightweight file watcher for developers. Monitors file changes and automatically re-runs your command — like [nodemon](https://github.com/remy/nodemon), but written in Go.

## Features

- **Recursive directory watching** — automatically picks up new subdirectories
- **Glob filtering** — include/exclude files with `**/*.go` style patterns
- **Debounce** — configurable delay to batch rapid changes into a single run
- **Process group kill** — cleanly terminates the entire process tree on restart
- **Colored log output** — clear visual distinction between change, exec, warn, and error events
- **Zero config** — sensible defaults, no config file required

## Install

```bash
go install github.com/chenyu695/gwatch@latest
```

Or build from source:

```bash
git clone https://github.com/chenyu695/gwatch.git
cd gwatch
go build -o gwatch .
```

## Usage

```
gwatch [flags] -x "command"
gwatch [flags] -- command args...
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-w <dir>` | `.` | Directory to watch (repeatable) |
| `-e <glob>` | `**/*` | Glob pattern to match (repeatable) |
| `-i <glob>` | — | Glob pattern to ignore (repeatable) |
| `-d <duration>` | `300ms` | Debounce delay |
| `-x <cmd>` | — | Shell command to execute on change |
| `-n` | `false` | Don't run the command on startup |

### Examples

Watch all Go files and rebuild on change:

```bash
gwatch -e "**/*.go" -x "go build -o server . && ./server"
```

Watch multiple directories, ignore test files:

```bash
gwatch -w ./src -w ./pkg -e "**/*.go" -i "**/*_test.go" -x "go test ./..."
```

Watch with a longer debounce for slow builds:

```bash
gwatch -e "**/*.go" -d 1s -x "make build"
```

Don't run on startup, use `--` to pass the command:

```bash
gwatch -n -e "**/*.js" -- npm run build
```

### Output

```
[23:14:48] [gwatch] Starting gwatch
[23:14:48] [gwatch] Watching:  .
[23:14:48] [gwatch] Patterns:  **/*.go
[23:14:48] [gwatch] Command:   go build -o server . && ./server
[23:14:48] [gwatch] Debounce:  300ms
[23:14:50] [change] main.go
[23:14:50] [exec]   $ go build -o server . && ./server
[23:14:51] [gwatch] Process finished
```

## How It Works

1. **Watcher** recursively registers all subdirectories with [fsnotify](https://github.com/fsnotify/fsnotify), filtering events through [doublestar](https://github.com/bmatcuk/doublestar) glob patterns
2. **Debounce** batches rapid file changes (e.g. editor saving multiple files) into a single trigger using `time.AfterFunc`
3. **Runner** executes the command via `sh -c`, using process groups (`setpgid`) so the entire process tree can be killed cleanly on restart

## License

[MIT](LICENSE)
