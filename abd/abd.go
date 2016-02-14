package abd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

const (
	DefaultConfigDir   = "/usr/lib/abd/sources.list.d/"
	DefaultStrategyDir = "/usr/lib/abd/strategies/"
)

func Discover(input, configDir, strategyDir string) (*schema.ABDMetadata, error) {
	// Parse the input into identifiers and labels. The identifier is needed
	// later, and this ensures we don't pass malformed input to a strategy.
	identifier, _, err := strategies.ParseInput(input)
	if err != nil {
		return nil, err
	}

	// load the configs from disk
	confs, err := getConfigs(configDir)
	if err != nil {
		return nil, err
	}

	for _, conf := range confs {
		// skip this strategy if the prefix doesn't match
		prefix := conf.ParsedConf.Prefix
		if !strings.HasPrefix(identifier, prefix) && prefix != "*" {
			continue
		}

		// check that the binary for this strategy exists
		binPath := path.Join(strategyDir, conf.ParsedConf.Strategy)
		_, err := os.Stat(binPath)
		if err != nil {
			return nil, err
		}

		// Run the strategy, passing in the provided input as an argument and
		// the configuration on stdin
		var out bytes.Buffer
		run := exec.Cmd{
			Path:   binPath,
			Args:   []string{binPath, input},
			Stdin:  bytes.NewBuffer(conf.RawConf),
			Stdout: &out,
			Stderr: os.Stderr,
		}
		err = run.Run()
		if err != nil {
			// If the error was with the strategy, return an error with its
			// exit code and error message.
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
				abdErr := &schema.ABDError{}
				err = json.Unmarshal(out.Bytes(), &abdErr)
				if err == nil {
					return nil, fmt.Errorf("exit code %d: %s", exitCode, abdErr.Error())
				}
			}
			return nil, err
		}

		metadata := &schema.ABDMetadata{}
		err = json.Unmarshal(out.Bytes(), metadata)
		if err != nil {
			return nil, err
		}
		return metadata, nil
	}
	// If we reach here, no strategies were found that had a prefix that
	// matched the given identifier.
	return nil, fmt.Errorf("no strategies found for identifier %s", identifier)
}

type ABDConfig struct {
	ParsedConf schema.ABDMetadataFetchStrategyConfiguration
	RawConf    []byte
}

// getConfigs will loop through the config directory and return the parsed
// version of all of the configs.
func getConfigs(configDir string) ([]ABDConfig, error) {
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var cfgs []ABDConfig

	for _, file := range files {
		f, err := os.Open(filepath.Join(configDir, file.Name()))
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		f.Close()

		var c schema.ABDMetadataFetchStrategyConfiguration
		err = json.Unmarshal(b, &c)
		if err != nil {
			// A bad file in the config directory (like vim's .swp files)
			// shouldn't prevent loading in the rest of the valid configs
			continue
		}
		cfgs = append(cfgs, ABDConfig{c, b})
	}

	if len(cfgs) == 0 {
		return nil, fmt.Errorf("no valid config files found")
	}

	return cfgs, nil
}

// BinaryName takes a name of an ABD discovery strategy, and returns the name
// of the binary implementing the strategy.
func BinaryName(strategy string) string {
	return "abd-" + strategy
}
