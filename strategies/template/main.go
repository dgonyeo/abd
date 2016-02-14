package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

type StrategyTemplateConfig struct {
	schema.ABDMetadataFetchStrategyConfiguration
	Template []schema.ABDMirror `json:"template"`
}

func main() {
	identifier, labels, confBlob := strategies.GetIdentifierLabelsAndConf()

	conf := &StrategyTemplateConfig{}
	err := json.Unmarshal(confBlob, conf)
	if err != nil {
		fmt.Println(strategies.MarshalError(err.Error()))
		os.Exit(strategies.DefaultErrorExitCode)
	}

	mirrors := make([]schema.ABDMirror, len(conf.Template))
	for i, m := range conf.Template {
		mirrors[i].Artifact = doTemplateSubstitution(m.Artifact, identifier, labels)
		mirrors[i].Signature = doTemplateSubstitution(m.Signature, identifier, labels)
	}

	metadata := &schema.ABDMetadata{
		Identifier: identifier,
		Labels:     labels,
		Mirrors:    mirrors,
	}

	metaBlob, err := json.Marshal(metadata)
	if err != nil {
		fmt.Println(strategies.MarshalError(err.Error()))
		os.Exit(strategies.DefaultErrorExitCode)
	}

	fmt.Print(string(metaBlob))
	os.Exit(0)
}

func doTemplateSubstitution(str string, identifier string, labels map[string]string) string {
	str = strings.Replace(str, "<identifier>", identifier, -1)
	for k, v := range labels {
		str = strings.Replace(str, "<"+k+">", v, -1)
	}
	return str
}
