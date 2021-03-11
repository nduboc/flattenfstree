package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const versionInfo = "v0.1"

const readDirBufferLen = 2

func main() {
	var rootCommand = &cobra.Command{
		Use:   "flattenfstree",
		Short: "flattenfstree moved all files of a given folder tree into a single target folder",

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
			err := moveFiles(args[0], target)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	if rootCommand.Execute() != nil {
		os.Exit(1)
	}
}

type stringSet map[string]struct{}

func moveFiles(source, target string) error {
	fmt.Printf("Moving files from %s into %s...\n", source, target)

	var copiedFiles stringSet
	copiedFiles, err := listDir(target)
	if err != nil {
		return err
	}
	fmt.Printf("%d files in target dir\n", len(copiedFiles))

	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Dir(path) == target {
				fmt.Printf("SKIP IN PLACE %s\n", path)
			} else {
				originalName := filepath.Base(path)
				destName := findAvailableName(originalName, copiedFiles)
				if originalName != destName {
					fmt.Printf("DUPLICATED %s to %s\n", path, destName)
				} else {
					fmt.Printf("MOVE %s to %s\n", path, destName)
				}
				copiedFiles[destName] = struct{}{}
			}
		}
		return nil
	})

	return err
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

func listDir(path string) (stringSet, error) {
	fileSet := make(stringSet)
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
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
