package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/appc/abd/abd"
)

const (
	exampleLocalStrategyConfiguration = `
	{
		"prefix": "*",
		"strategy": "io.abd.local",
		"storage-path": "/var/abd"
	}
	`
)

var (
	configDir   = abd.DefaultConfigDir
	strategyDir = abd.DefaultStrategyDir
	abdCmd      = &cobra.Command{
		Use:   "abd",
		Short: "abd - the appc Binary Discovery",
		Long:  `abd is a framework for resolving human-readable strings to downloadable URIs`,
	}
	discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "discover an artefact using ABD",
		Run:   discoverFunc,
	}
	mirrorsCmd = &cobra.Command{
		Use:   "mirrors",
		Short: "discover the mirrors for an artefact using ABD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	fetchCmd = &cobra.Command{
		Use:   "fetch",
		Short: "discover and fetch an artefact using ABD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
)

func init() {
	abdCmd.PersistentFlags().StringVarP(&configDir, "config-dir", "", "", "configuration directory for abd")
	abdCmd.PersistentFlags().StringVarP(&strategyDir, "strategy-dir", "", "", "strategy directory for abd")
	abdCmd.AddCommand(discoverCmd, mirrorsCmd, fetchCmd)
}

func main() {
	abdCmd.Execute()
}

// abd discover identifier,label1=value1,label2=value2,...
func discoverFunc(cmd *cobra.Command, args []string) {
	if len(args) != 1 || args[0] == "" {
		fmt.Println("no identifier given")
		os.Exit(1)
	}
	metadata, err := abd.Discover(args[0], configDir, strategyDir)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	formattedMetadata, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(formattedMetadata))
}
