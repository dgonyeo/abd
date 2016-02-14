// Common strategies code, defining the plugin execution interface
package strategies

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/appc/abd/schema"
)

const (
	// The default exit code used when an error is encountered
	DefaultErrorExitCode = 1
)

// ParseInput takes a string containing an identifier and labels, and returns
// the identifier and a map with all of the labels. If any duplicate label
// names are encountered, an error is returned.
func ParseInput(input string) (string, map[string]string, error) {
	// The identifier and labels is separated by a comma
	parts := strings.Split(input, ",")
	identifier := parts[0]
	if identifier == "" {
		return "", nil, fmt.Errorf("no identifier given")
	}

	// The labels come after the initial comma, and are separated by commas.
	// Each label has a name and a value, which are separated by the first =
	labels := make(map[string]string)
	for _, part := range parts[1:] {
		index := strings.Index(part, "=")
		if index == -1 {
			return "", nil, fmt.Errorf("invalid label %s", part)
		}
		name := part[:index]
		value := part[index+1:]

		if _, ok := labels[name]; ok {
			return "", nil, fmt.Errorf("label %s defined twice", name)
		}

		labels[name] = value
	}
	return identifier, labels, nil
}

// MarshalError will convert a string (presumably containing an error message)
// into a marshalled schema.ABDError with the provided message.
func MarshalError(str string) string {
	abdErr := &schema.ABDError{E: str}
	out, err := json.Marshal(abdErr)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// GetIdentifierLabelsAndConf will obtain the identifier, labels, and
// configuration from the program's arguments and stdin. If an error are
// encountered, the error is printed to stdout and the program exits with the
// DefaultErrorExitCode.
func GetIdentifierLabelsAndConf() (string, map[string]string, []byte) {
	// Get the identifier and labels from the arguments
	// args[0] = abd-{strategy}
	// args[1] = {identifier},{labels}
	args := os.Args
	if len(args) != 2 {
		fmt.Println(MarshalError(fmt.Sprintf("incorrect number of args: %v", args)))
		os.Exit(DefaultErrorExitCode)
	}
	// Parse out the identifier and labels
	identifier, labels, err := ParseInput(args[1])
	if err != nil {
		fmt.Println(MarshalError(err.Error()))
		os.Exit(DefaultErrorExitCode)
	}
	// Read the configuration from stdin
	conf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(MarshalError(err.Error()))
		os.Exit(DefaultErrorExitCode)
	}
	return identifier, labels, conf
}
