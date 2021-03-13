// Copyright (C) 2021 Nicolas Duboc
// MIT License

package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const versionInfo = "v0.1"

const readDirBufferLen = 100

func main() {
	var apply bool
	var rootCommand = &cobra.Command{
		Use:   "flattenfstree [flags] sourceDir [targetDir]",
		Short: "flattenfstree moves all files of sourceDir and its sub-dirs into a single \ntarget directory.",
		Long: `flattenfstree moves all files of sourceDir and its sub-dirs into a single
target directory.

If no target directory is given, the files of the subdirectories of the
source are moved to the source directory itself  (target directory is the
source directory).
				 
Collisions on file names are handled by adding a numeric count to the
filenames: '...-1.jpg', '...-2.jpg', ...

Sub-directories of the source are deleted after all files have been moved.

By default, the command only shows what will be performed and doesn't move
files or delete directories.  Use the --apply flag to actually perform the
moves.`,

		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.RangeArgs(1, 2)(cmd, args)
			if err != nil {
				return err
			}
			srcInfo, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if !srcInfo.IsDir() {
				return fmt.Errorf("path is not a directory: %s", args[0])
			}
			if len(args) > 1 {
				targetInfo, err := os.Stat(args[1])
				if err != nil {
					return err
				}
				if !targetInfo.IsDir() {
					return fmt.Errorf("path is not a directory: %s", args[1])
				}
				absSource, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}
				absTarget, err := filepath.Abs(args[1])
				if err != nil {
					return err
				}
				if absTarget != absSource && strings.HasPrefix(absTarget, absSource) {
					return fmt.Errorf("target folder is inside the source folder")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var target string
			if len(args) > 1 {
				target = args[1]
			} else {
				target = args[0]
			}
			initialCount, copiedCount, err := moveFiles(args[0], target, apply)
			fmt.Printf("%d files and directories initially in target folder\n", initialCount)
			if apply {
				fmt.Printf("%d files moved from source directory\n", copiedCount)
			} else {
				fmt.Printf("%d files to be moved from source directory\n", copiedCount)
				fmt.Printf("No file was moved (no --apply flag)\n")
			}

			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
		},
	}
	rootCommand.PersistentFlags().BoolVar(&apply, "apply", false,
		"Do move files instead of simplify showing what will be done")

	if rootCommand.Execute() != nil {
		os.Exit(1)
	}
}

type stringSet map[string]struct{}

func moveFiles(source, target string, apply bool) (int, int, error) {
	fmt.Printf("Moving files from %s into %s...\n", source, target)

	var targetEntries stringSet
	targetEntries, err := listDir(target)
	if err != nil {
		return 0, 0, err
	}
	initialCount := len(targetEntries)
	copiedCount := 0

	// emptyDirs will collect traversed dirs to be deleted after files have been moved
	emptyDirs := make([]string, 0, 10)

	sourceStat, err := os.Stat(source)
	if err != nil {
		return 0, 0, err
	}
	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			if filepath.Dir(path) == target {
				fmt.Printf("SKIP IN PLACE %s\n", path)
			} else {
				originalName := d.Name()
				destName := findAvailableName(originalName, targetEntries)
				if originalName != destName {
					fmt.Printf("DUPLICATED %s to %s\n", path, destName)
				} else {
					fmt.Printf("MOVE %s to %s\n", path, destName)
				}
				if apply {
					info, err := d.Info()
					if err != nil {
						return err
					}
					err = doMoveFile(path, filepath.Join(target, destName), info.ModTime())
					if err != nil {
						return err
					}
				}
				targetEntries[destName] = struct{}{}
				copiedCount++
			}
		} else {
			pathInfo, err := d.Info()
			if err != nil {
				return err
			}
			if !os.SameFile(sourceStat, pathInfo) {
				emptyDirs = append(emptyDirs, path)
			}
		}
		return nil
	})
	fmt.Println()

	if err == nil && apply {
		for i := len(emptyDirs) - 1; err == nil && i >= 0; i-- {
			err = os.Remove(emptyDirs[i])
			if err == nil {
				fmt.Printf("DELETE DIR %s\n", emptyDirs[i])
			}
		}
		if err != nil {
			err = fmt.Errorf("error while deleting directory: %s", err)
		}
		fmt.Println()
	}

	return initialCount, copiedCount, err
}

func findAvailableName(original string, files stringSet) string {
	// assume we'll find a candidate before i overflows
	i := 1
	candidate := original
	for {
		_, ok := files[candidate]
		if !ok {
			break // found an unused name
		}
		candidate = injectInt(original, i)
		i++
	}
	return candidate
}

func injectInt(filename string, i int) string {
	mainNameStart := 0 // index after trailing dots
	for ; filename[mainNameStart] == '.' && mainNameStart < len(filename); mainNameStart++ {
	}
	if mainNameStart == len(filename) {
		panic(fmt.Sprintf("Unsupported filename '%s'", filename))
	}
	mainPart := filename[mainNameStart:]
	ext := filepath.Ext(mainPart)
	return fmt.Sprintf("%s%s-%d%s", filename[0:mainNameStart], strings.TrimSuffix(mainPart, ext), i, ext)
}

// listDir return names of files and directories in the given folder
func listDir(path string) (stringSet, error) {
	fileSet := make(stringSet)
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	for {
		list, err := dir.Readdirnames(readDirBufferLen)
		if len(list) == 0 {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		for _, n := range list {
			fileSet[n] = struct{}{}
		}
	}

	return fileSet, nil
}

func doMoveFile(source, dest string, mtime time.Time) error {
	err := os.Rename(source, dest)
	if err != nil {
		return err
	}
	err = os.Chtimes(dest, mtime, mtime)
	if err != nil {
		return err
	}
	return nil
}
