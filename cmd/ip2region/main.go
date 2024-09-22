package main

import (
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.AddCommand(createQueryCommand(), createWebCommand(), createUpdateCommand())
	root.InitDefaultHelpCmd()
	for _, c := range root.Commands() {
		if c.Name() == "help" {
			c.Short = "显示帮助"
		}
	}

	root.Execute()
}
