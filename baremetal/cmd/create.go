/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/aojea/kind-networking-plugins/pkg/docker"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// Config struct for multicluster config
type Config struct {
	Cluster  v1alpha4.Cluster `yaml:"cluster"`
	Networks []string         `yaml:"networks"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a baremetal cluster",
	Long: `Create a baremetal cluster.

Create multiple network and interfaces on the nodes, that may
be used to have dedicated network for storage, external servoces, ...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createBareMetal(cmd)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().String(
		"name",
		cluster.DefaultName,
		"the multicluster context name",
	)
	createCmd.Flags().String(
		"config",
		"./config.yml",
		"the config file with the cluster configuration",
	)
	createCmd.MarkFlagRequired("config")
}

func createBareMetal(cmd *cobra.Command) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	cfg, err := NewConfig(configPath)
	if err != nil {
		return err
	}

	// create the clusters
	logger := kindcmd.NewLogger()
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)
	clusterNetwork := "bm-" + name
	// use a separate network for the cluster
	// autoallocate subnet and allow masquerading
	err = docker.CreateNetwork(clusterNetwork, "", true)
	if err != nil {
		return err
	}

	// use the new created docker network
	os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", clusterNetwork)

	// create the cluster
	if err := provider.Create(
		name,
		cluster.CreateWithV1Alpha4Config(&cfg.Cluster),
		// cluster.CreateWithNodeImage(flags.ImageName),
		// cluster.CreateWithRetain(flags.Retain),
		// cluster.CreateWithWaitForReady(flags.Wait),
		// cluster.CreateWithKubeconfigPath(flags.Kubeconfig),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return errors.Wrap(err, "failed to create cluster")
	}
	// reset the env variable
	os.Unsetenv("KIND_EXPERIMENTAL_DOCKER_NETWORK")
	// create the secondary interfaces in all nodes
	nodes, err := provider.ListNodes(name)
	if err != nil {
		return err
	}
	for _, networkName := range cfg.Networks {
		err = docker.CreateNetwork(networkName, "", false)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			err := docker.ConnectNetwork(n.String(), networkName, "")
			if err != nil {
				return err
			}
		}
	}

	return nil
}
