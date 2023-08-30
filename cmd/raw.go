/*
Copyright © 2023 Dataflows
*/
package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/thedataflows/go-commons/pkg/config"
	"github.com/thedataflows/go-commons/pkg/file"
	"github.com/thedataflows/go-commons/pkg/log"

	"github.com/spf13/cobra"
)

type Command struct {
	Module     string   `yaml:"module"`
	Controller string   `yaml:"controller"`
	Command    string   `yaml:"command"`
	Method     string   `yaml:"method"`
	Parameters []string `yaml:"parameters,omitempty"`
}

var (
	cmdRawCommand = &cobra.Command{
		Use:     "raw",
		Short:   "Call OPNSense Raw API command",
		Long:    ``,
		Aliases: []string{"r"},
		Run:     RunRawCommand,
	}
)

func init() {
	rootCmd.AddCommand(cmdRawCommand)

	var commandsFile string
	// Set persistent flags instead of local flags to be able to use them in subcommands
	cmdRawCommand.PersistentFlags().StringVar(&commandsFile, "commands-file", "raw-commands.yaml", "Commands file")
	// force parsing of flags
	_ = cmdRawCommand.ParseFlags(os.Args)

	config.ViperBindPFlagSet(cmdRawCommand, cmdRawCommand.PersistentFlags())

	if !file.IsAccessible(commandsFile) {
		log.Fatalf("Config file %s is not accessible", commandsFile)
	}

	contents, err := os.ReadFile(commandsFile)
	if err != nil {
		log.Fatal(err)
	}

	var rawCommands []Command
	err = yaml.UnmarshalWithOptions(contents, &rawCommands, yaml.Strict())
	if err != nil {
		log.Fatalf("Failed to parse file %s: %v", commandsFile, err)
	}

	apiCategory := filepath.Base(commandsFile)
	apiCategory = apiCategory[:len(apiCategory)-len(filepath.Ext(apiCategory))]

	// Add subcommands
	for _, subcommand := range rawCommands {
		group := &cobra.Group{
			ID:    fmt.Sprintf("%s/%s", subcommand.Module, subcommand.Controller),
			Title: fmt.Sprintf("Module: %s, Controller: %s", subcommand.Module, subcommand.Controller),
		}
		if !cmdRawCommand.ContainsGroup(group.ID) {
			cmdRawCommand.AddGroup(group)
		}

		short := fmt.Sprintf("Method: %s", subcommand.Method)
		if len(subcommand.Parameters) > 0 {
			short = fmt.Sprintf("%s, Arguments: %s", short, subcommand.Parameters)
		}
		subCmd := &cobra.Command{
			Use:     fmt.Sprintf("%s/%s/%s", subcommand.Module, subcommand.Controller, subcommand.Command),
			GroupID: group.ID,
			Short:   short,
			Args:    cobra.MinimumNArgs(len(subcommand.Parameters)),
			Annotations: map[string]string{
				"method": subcommand.Method,
			},
			Long: fmt.Sprintf("\nhttps://docs.opnsense.org/development/api/%s/%s.html\n\n%s", apiCategory, subcommand.Module, short),
			Run: func(cmd *cobra.Command, args []string) {
				callingURL := fmt.Sprintf("%s/api/%s", config.ViperGetString(cmd.Root(), keyCommonOpnSenseURL), cmd.Use)
				if len(args) > 0 {
					callingURL = fmt.Sprintf("%s/%s", callingURL, strings.Join(args, "/"))
				}
				opnsenseKey := config.ViperGetString(cmd.Root(), keyCommonOpnSenseKey)
				opnsenseSecret := config.ViperGetString(cmd.Root(), keyCommonOpnSenseSecret)
				opnsenseSecretFile := config.ViperGetString(cmd.Root(), keyCommonOpnSenseSecretFile)
				if (len(opnsenseKey) == 0 || len(opnsenseSecret) == 0) && len(opnsenseSecretFile) > 0 {
					contents, err := os.ReadFile(opnsenseSecretFile)
					if err != nil {
						log.Fatal(err)
					}
					lines := strings.Split(string(contents), "\n")
					if len(lines) < 2 {
						log.Fatalf("Key/Secret file %s must contain key= and secret= fields", opnsenseSecretFile)
					}
					for _, l := range lines {
						if strings.HasPrefix(l, "key=") {
							opnsenseKey = l[4:]
						}
						if strings.HasPrefix(l, "secret=") {
							opnsenseSecret = l[7:]
						}
					}
				}
				log.Debugf("Key: %s, Secret: %s", opnsenseKey, opnsenseSecret)

				callOpnSenseAPI(
					callingURL,
					cmd.Annotations["method"],
					opnsenseKey,
					opnsenseSecret,
					config.ViperGetBool(cmd.Root(), keyCommonOpnSenseURLInsecure),
				)
			},
		}
		cmdRawCommand.AddCommand(subCmd)
	}
}

func RunRawCommand(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}

func callOpnSenseAPI(url string, method string, key string, secret string, insecure bool) {
	log.Infof("%s %s", method, url)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}} // #nosec

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %s", err)
	}

	req.SetBasicAuth(key, secret)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error calling %s: %s", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %s", err)
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("Error parsing response body: %s\n%s", err, body)
	}
	prettyJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Fatalf("Error formatting response body: %s", err)
	}

	fmt.Println(string(prettyJSON))
}
