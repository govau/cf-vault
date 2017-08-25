package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

const (
	orgPrefix      = "cf_o/"
	spacePrefix    = "cf_s/"
	instancePrefix = "cf_i/"
)

type cfVault struct{}

type serviceKeys struct {
	Resources []struct {
		Entity struct {
			Name        string `json:"name"`
			Credentials struct {
				Address string `json:"Address"`
				Auth    struct {
					Token string `json:"token"`
				} `json:"auth"`
				Backends struct {
					Generic string `json:"generic"`
				} `json:"backends"`
				SharedBackends struct {
					Organization string `json:"organization"`
					Space        string `json:"space"`
				} `json:"backends_shared"`
			} `json:"credentials"`
		} `json:"entity"`
	} `json:"resources"`
}

func (c *cfVault) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "vault" {
		if len(args) < 2 {
			log.Fatal("need at least one-arg, name of the vault instance. Create with cf create-service hashicorp-vault shared my-vault")
		}

		serviceName := args[1]
		service, err := cliConnection.GetService(serviceName)
		if err != nil {
			log.Fatal("error getting service: ", err)
		}

		at, err := cliConnection.AccessToken()
		if err != nil {
			log.Fatal("error getting user access token: ", err)
		}

		apiEndpoint, err := cliConnection.ApiEndpoint()
		if err != nil {
			log.Fatal("error getting API endpoint: ", err)
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/service_instances/%s/service_keys", apiEndpoint, url.PathEscape(service.Guid)), nil)
		if err != nil {
			log.Fatal("error creating HTTP request: ", err)
		}
		req.Header.Add("Authorization", at)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal("error creating fetching service keys: ", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal("did not get 200 status back:", resp.Status)
		}

		var sks serviceKeys
		err = json.NewDecoder(resp.Body).Decode(&sks)
		if err != nil {
			log.Fatal("error decoding response for keys:", err)
		}

		if len(sks.Resources) == 0 {
			log.Fatalf("no service keys found. create one with: cf create-service-key %s my-key", serviceName)
		}

		skE := sks.Resources[0].Entity

		log.Printf("Using Vault instance: %s with service key: %s\n", serviceName, skE.Name)

		var argsToSend []string
		for _, a := range args[2:] { // skip vault my-service
			// Look for args beginning with a prefix, and substitute in appropriate values
			switch {
			case strings.HasPrefix(a, orgPrefix):
				a = fmt.Sprintf("%s/%s", skE.Credentials.SharedBackends.Organization, a[len(orgPrefix):])
			case strings.HasPrefix(a, spacePrefix):
				a = fmt.Sprintf("%s/%s", skE.Credentials.SharedBackends.Space, a[len(spacePrefix):])
			case strings.HasPrefix(a, instancePrefix):
				a = fmt.Sprintf("%s/%s", skE.Credentials.Backends.Generic, a[len(instancePrefix):])
			}
			argsToSend = append(argsToSend, a)
		}

		cmd := exec.Command("vault", argsToSend...)
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("VAULT_TOKEN=%s", skE.Credentials.Auth.Token),
			fmt.Sprintf("VAULT_ADDR=%s", skE.Credentials.Address),
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			log.Fatal("error running Vault command:", err)
		}
	}
}

func (c *cfVault) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Plugin to make it easy to work with the service broker provided by Hashicorp Vault",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "vault",
				HelpText: "vault, subcommand - automatically logged in to space.",
				UsageDetails: plugin.Usage{
					Usage: "vault\n   cf vault my-service-name read {cf_o,cf_s,cf_i}/xxx",
				},
			},
		},
	}
}

func main() {
	plugin.Start(&cfVault{})
}
