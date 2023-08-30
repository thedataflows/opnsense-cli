/*
Copyright © 2023 Dataflows
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thedataflows/go-commons/pkg/log"
)

var (
	cmdMacroRun = &cobra.Command{
		Use:     "run",
		Short:   "Run a predefined macro",
		Long:    ``,
		Aliases: []string{"r"},
		Run:     RunMacroRun,
	}
)

func init() {
	cmdMacro.AddCommand(cmdMacroRun)
}

func RunMacroRun(_ *cobra.Command, args []string) {
	macroList := loadMacroFile(macroFileName)

	if len(args) == 0 {
		log.Error("No macro names specified, select at least one from the list")
		cmdMacroList.Run(cmdMacroList, args)
		return
	}

	found := false
	for _, macro := range *macroList {
		if macro.Name == args[0] {
			found = true
			log.Infof("Running macro '%s'", macro.Name)
			for _, command := range macro.Commands {
				// log.Infof("Running command '%s'", command)
				for _, cmd := range cmdRawCommand.Commands() {
					if cmd.Name() == command {
						cmd.Run(cmd, args[1:])
						break
					}
				}
			}
		}
	}

	if !found {
		log.Warnf("Macro '%s' not found", args[0])
	}

}
