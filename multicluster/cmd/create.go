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
	"fmt"
	"os"

	"github.com/aojea/kind-networking-plugins/pkg/docker"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/exec"
)

const dockerWanImage = "quay.io/aojea/wanem:latest"

const rawConfig = `
# three node (two workers) cluster config
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ipv4
nodes:
- role: control-plane
`

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a deployment with multiple KIND clusters",
	Long: `Create a deployment with multiple KIND clusters based on the configuration
	passed as parameters.
	
	Multicluster deployment create KIND clusters in independent bridges, that are connected
	through an special container that handles the routing and the WAN emulation.
	`,
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
	createCmd.Flags().Int(
		"number",
		2,
		"the number of clusters (default 2)",
	)
}

func configureMultiCluster(cmd *cobra.Command) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	number, err := cmd.Flags().GetInt("number")
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

	for i := 0; i < number; i++ {
		clusterName := fmt.Sprintf("multi-%s-%d", name, i)
		// each cluster has its own docker network with the clustername
		// TODO: remove this hack to create custom subnets
		subnet := fmt.Sprintf("10.10.%d.0/24", i)
		gateway := fmt.Sprintf("10.10.%d.254", i)
		err := docker.CreateNetwork(clusterName, subnet, false)
		if err != nil {
			return err
		}
		err = docker.ConnectNetwork(wanem, clusterName, gateway)
		if err != nil {
			return err
		}
		// use the new created docker network
		os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", clusterName)
		podSubnet := fmt.Sprintf("10.10%d.0.0/16", i)
		svcSubnet := fmt.Sprintf("10.96.%d.0/24", i)
		config := &v1alpha4.Cluster{
			Name: clusterName,
			Nodes: []v1alpha4.Node{
				{
					Role: v1alpha4.ControlPlaneRole,
				},
			},
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
			err := docker.ReplaceGateway(n.String(), gateway)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createWanem(name string) error {
	containerName := "wan-" + name
	args := []string{"run",
		"-d",                    // run in the background
		"--name", containerName, // well known name
		dockerWanImage,
	}

	cmd := exec.Command("docker", args...)
	return cmd.Run()

}
