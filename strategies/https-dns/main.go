package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

var (
	errGoUp = fmt.Errorf("didn't find any metadata, go up the path")
)

func main() {
	identifier, labels, _ := strategies.GetIdentifierLabelsAndConf()

	for {
		metadata, err := getMetadata(identifier, labels)
		if err == errGoUp {
			newidentifier, _ := path.Split(identifier)
			if newidentifier == identifier {
				break
			}
			identifier = newidentifier
			continue
		}
		if err != nil {
			fmt.Println(strategies.MarshalError(err.Error()))
			os.Exit(strategies.DefaultErrorExitCode)
		}

		out, err := json.Marshal(metadata)
		if err != nil {
			fmt.Println(strategies.MarshalError(err.Error()))
			os.Exit(strategies.DefaultErrorExitCode)
		}
		fmt.Println(string(out))
		os.Exit(0)
	}
	fmt.Print(strategies.MarshalError("couldn't find artifact"))
	os.Exit(strategies.DefaultErrorExitCode)
}

func getMetadata(identifier string, labels map[string]string) ([]schema.ABDMetadata, error) {
	index := strings.Index(identifier, "/")

	var domain, location string

	if index == -1 {
		domain = identifier
		location = ""
	} else {
		domain = identifier[:index]
		location = identifier[index+1:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://"+path.Join(domain, ".abd", location), nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	for k, v := range labels {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errGoUp
	}

	metablob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	metadata := []schema.ABDMetadata{}
	err = json.Unmarshal(metablob, &metadata)
	if err != nil {
		return nil, errGoUp
	}

	return metadata, nil
}
