package main

import (
	"fmt"
	"os"

	"github.com/appc/abd/strategies"
)

func main() {
	identifier, _, _ := strategies.GetIdentifierLabelsAndConf()
	fmt.Println(strategies.MarshalError(fmt.Sprintf("refusing to discover %q, configured to be a noop", identifier)))
	os.Exit(strategies.DefaultErrorExitCode)
}
