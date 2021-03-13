# flattenfstree

`flattenfstree` is a command-line utility to move all files of sourceDir and its sub-dirs into a single
target directory.

```
Usage:
  flattenfstree [flags] sourceDir [targetDir]

If no target directory is given, the files of the subdirectories of the
source are moved to the source directory itself  (target directory is the
source directory).
				 
Collisions on file names are handled by adding a numeric count to the
filenames: '...-1.jpg', '...-2.jpg', ...

Sub-directories of the source are deleted after all files have been moved.

By default, the command only shows what will be performed and doesn't move
files or delete directories.  Use the --apply flag to actually perform the
moves.

Usage:
  flattenfstree [flags] sourceDir [targetDir]

Flags:
      --apply   Do move files instead of simplify showing what will be done
  -h, --help    help for flattenfstree
```

Tested on macos and Windows 10.

## Build

Requires Go 1.16.

Use `build.sh` script or simply `go build`

## Licence

Copyright (c) 2021 Nicolas Duboc
MIT License
