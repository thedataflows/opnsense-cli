/*
Copyright © 2023 Dataflows
*/
package cmd

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/thedataflows/go-commons/pkg/config"
	"github.com/thedataflows/go-commons/pkg/log"
)

const (
	keyCmdMacroFile = "macro-file"
)

var (
	cmdMacro = &cobra.Command{
		Use:     "macro",
		Short:   "Run predefined macros",
		Long:    ``,
		Aliases: []string{"m"},
		Run:     RunMacro,
	}

	macroFileName string
)

type Macro struct {
	Name     string   `yaml:"name"`
	Commands []string `yaml:"commands"`
}

func init() {
	rootCmd.AddCommand(cmdMacro)

	cmdMacro.PersistentFlags().StringVar(&macroFileName, keyCmdMacroFile, "default-macro.yaml", "Macro file, YAML format")

	config.ViperBindPFlagSet(cmdMacro, cmdMacro.PersistentFlags())
}

func loadMacroFile(mFile string) *[]Macro {
	contents, err := os.ReadFile(mFile)
	if err != nil {
		log.Fatal(err)
	}

	var macroList []Macro
	err = yaml.UnmarshalWithOptions(contents, &macroList, yaml.Strict())
	if err != nil {
		log.Fatalf("Failed to parse file %s: %v", mFile, err)
	}

	return &macroList
}

func RunMacro(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}
