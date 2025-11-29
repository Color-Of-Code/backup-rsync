package main

import (
	"backup-rsync/backup/cmd"
	"os"
)

func main() {
	rootCmd := cmd.BuildRootCommand()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
