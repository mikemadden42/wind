# wind
Find directory size in Go

`wind` prints the total size of each subdirectory of a directory, one per
line, in a format that pipes cleanly into `sort -n`.

## Install

```sh
go install github.com/mikemadden42/wind@latest
```

## Usage

```sh
wind [-dir DIR] [-unit B|KB|MB]
```

| Flag    | Default | Meaning                              |
|---------|---------|--------------------------------------|
| `-dir`  | `.`     | Directory whose subdirectories to size |
| `-unit` | `MB`    | Output unit: `B`, `KB`, or `MB`      |

```console
$ wind -dir ~/src -unit KB | sort -n
12	notes
408	wind
91320	website
```

## Notes

- Sizes are apparent file sizes (what `ls -l` shows), not blocks allocated on
  disk, so sparse files count at full size and hard-linked files count once per
  link. Symlinks count as their own size and are not followed.
- Sizes are rounded up to the next whole unit, as `du` does, so a non-empty
  directory never shows `0`.
- Only subdirectories are listed; files directly inside `-dir` are ignored.
- Errors go to stderr. Unreadable entries are skipped; if a whole subdirectory
  can't be read it is left out and `wind` exits with status 1.
