/*
Copyright © 2023 Dataflows
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/iancoleman/strcase"
	"github.com/thedataflows/go-commons/pkg/config"
	"github.com/thedataflows/opnsense-cli/pkg/constants"

	"github.com/spf13/cobra"
)

const (
	keyCommonOpnSenseURL         = "opnsense-url"
	keyCommonOpnSenseKey         = "opnsense-key"
	keyCommonOpnSenseSecret      = "opnsense-secret"
	keyCommonOpnSenseURLInsecure = "opnsense-url-insecure"
	keyCommonOpnSenseSecretFile  = "opnsense-secret-file"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "opnsense-cli",
		Short: "OPNSense command line interface",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Long = fmt.Sprintf(
				"%s\n\nAll flags values can be provided via env vars starting with %s_*\nExamples:\n- %s_%s=opnsenseapikeyname\n- %s_%s=opnsenseapisecret\nTo pass a command (e.g. 'command1') flag, use %s_COMMAND1_FLAGNAME=somevalue",
				cmd.Short,
				configOpts.EnvPrefix,
				configOpts.EnvPrefix,
				strcase.ToScreamingSnake(keyCommonOpnSenseKey),
				configOpts.EnvPrefix,
				strcase.ToScreamingSnake(keyCommonOpnSenseSecret),
				configOpts.EnvPrefix,
			)

			_ = cmd.Help()
		},
	}

	configOpts = config.DefaultConfigOpts(
		&config.Opts{
			EnvPrefix: constants.ViperEnvPrefix,
		},
	)
)

func initConfig() {
	config.InitConfig(configOpts)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().AddFlagSet(configOpts.Flags)
	rootCmd.PersistentFlags().String(keyCommonOpnSenseURL, "https://opnsense.local", "OPNSense URL")
	rootCmd.PersistentFlags().String(keyCommonOpnSenseKey, "", "OPNSense Key. See https://docs.opnsense.org/development/api.html#introduction")
	rootCmd.PersistentFlags().String(keyCommonOpnSenseSecret, "", "OPNSense Secret")
	rootCmd.PersistentFlags().String(keyCommonOpnSenseSecretFile, "", "Optional OPNSense Key and Secret File (downloaded from gui)")
	rootCmd.PersistentFlags().Bool(keyCommonOpnSenseURLInsecure, false, "OPNSense URL is Insecure")

	config.ViperBindPFlagSet(rootCmd, rootCmd.PersistentFlags())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
