/*
Copyright © 2023 Dataflows
*/
package cmd

import (
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/thedataflows/go-commons/pkg/log"
)

var (
	cmdMacroList = &cobra.Command{
		Use:     "list",
		Short:   "List predefined macros",
		Long:    ``,
		Aliases: []string{"l"},
		Run:     RunMacroList,
	}
)

func init() {
	cmdMacro.AddCommand(cmdMacroList)
}

func RunMacroList(cmd *cobra.Command, _ []string) {
	macroList := loadMacroFile(macroFileName)

	m, err := yaml.MarshalWithOptions(macroList, yaml.Indent(2))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("%s %s:\n%s", cmd.Parent().Name(), cmd.Name(), m)
}
