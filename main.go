package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const versionInfo = "v0.1"

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

func moveFiles(source, target string) error {
	fmt.Printf("Moving %s into %s\n", source, target)
	return nil
}
