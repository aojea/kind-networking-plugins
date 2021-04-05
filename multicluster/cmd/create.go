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
	"github.com/aojea/kind-networking-plugins/pkg/network"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/exec"
)

const dockerWanImage = "quay.io/aojea/wanem:latest"

// Config struct for multicluster config
type Config struct {
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	Nodes         int    `yaml:"nodes"`
	NodeSubnet    string `yaml:"nodeSubnet"`
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
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
	Short: "Create a deployment with multiple KIND clusters",
	Long: `Create a deployment with multiple KIND clusters based on the configuration
passed as parameters.

Multicluster deployment create KIND clusters in independent bridges, that are connected
through an special container that handles the routing and the WAN emulation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configureMultiCluster(cmd)
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

func configureMultiCluster(cmd *cobra.Command) error {
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

	// create the container to emulate the WAN network
	wanem := "wan-" + name
	err = createWanem(name)
	if err != nil {
		return err
	}

	// create the clusters
	logger := kindcmd.NewLogger()
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	for clusterName, clusterConfig := range cfg.Clusters {
		// each cluster has its own docker network with the clustername
		subnet := clusterConfig.NodeSubnet
		err := docker.CreateNetwork(clusterName, subnet, false)
		if err != nil {
			return err
		}
		// connect wanem with the last IP of the range
		// that the cluster will use later as gateway
		gateway, err := network.GetLastIPSubnet(subnet)
		if err != nil {
			return err
		}
		err = docker.ConnectNetwork(wanem, clusterName, gateway.String())
		if err != nil {
			return err
		}
		// use the new created docker network
		os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", clusterName)
		podSubnet := clusterConfig.PodSubnet
		svcSubnet := clusterConfig.ServiceSubnet
		config := &v1alpha4.Cluster{
			Name:  clusterName,
			Nodes: createNodes(clusterConfig.Nodes),
			Networking: v1alpha4.Networking{
				PodSubnet:     podSubnet,
				ServiceSubnet: svcSubnet,
			},
		}

		// create the cluster
		if err := provider.Create(
			clusterName,
			cluster.CreateWithV1Alpha4Config(config),
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
		// change the default network in all nodes
		// to use the wanem container and provide
		// connectivity between clusters
		nodes, err := provider.ListNodes(clusterName)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			err := docker.ReplaceGateway(n.String(), gateway.String())
			if err != nil {
				return err
			}
		}
		// insert routes in wanem to reach services through one of the nodes
		ipv4, _, err := nodes[0].IP()
		if err != nil {
			return err
		}
		err = addRoutesWanem(name, ipv4, svcSubnet, podSubnet)
		if err != nil {
			return err
		}

	}
	return nil
}

func createWanem(name string) error {
	containerName := "wan-" + name
	args := []string{"run",
		"-d", // run in the background
		"--sysctl=net.ipv4.ip_forward=1",
		"--sysctl=net.ipv4.conf.all.rp_filter=0",
		"--privileged",
		"--name", containerName, // well known name
		dockerWanImage,
	}

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}

func addRoutesWanem(name, gateway string, subnets ...string) error {
	for _, subnet := range subnets {
		args := []string{"exec", "wan-" + name,
			"ip", "route", "add", subnet, "via", gateway,
		}
		cmd := exec.Command("docker", args...)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func createNodes(n int) []v1alpha4.Node {
	nodes := []v1alpha4.Node{
		{
			Role: v1alpha4.ControlPlaneRole,
		},
	}

	for j := 1; j < n; j++ {
		n := v1alpha4.Node{
			Role: v1alpha4.WorkerRole,
		}
		nodes = append(nodes, n)
	}
	return nodes
}
