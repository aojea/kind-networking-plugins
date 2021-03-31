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
)

const topologyLabel = "topology.kubernetes.io/zone"

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
	createCmd.Flags().Int(
		"interfaces",
		2,
		"the number of interfaces per node (default 2)",
	)
	createCmd.Flags().Int(
		"nodes",
		2,
		"the number of nodes (default 2)",
	)
}

func createBareMetal(cmd *cobra.Command) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	interfaces, err := cmd.Flags().GetInt("interfaces")
	if err != nil {
		return err
	}
	nNodes, err := cmd.Flags().GetInt("nodes")
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

	config := &v1alpha4.Cluster{
		Name:  name,
		Nodes: createNodes(nNodes),
	}

	// create the cluster
	if err := provider.Create(
		name,
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
	// create the secondary interfaces in all nodes
	nodes, err := provider.ListNodes(name)
	if err != nil {
		return err
	}
	for i := 1; i < interfaces; i++ {
		networkName := fmt.Sprintf("%s-%d", clusterNetwork, i)
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

func createNodes(n int) []v1alpha4.Node {
	nodes := []v1alpha4.Node{
		{
			Role: v1alpha4.ControlPlaneRole,
		},
	}
	// TODO we use only one control plane per zone
	// because we are interested on the workers nodes by now
	for i := 1; i < n; i++ {
		node := v1alpha4.Node{
			Role: v1alpha4.WorkerRole,
		}
		nodes = append(nodes, node)
	}
	return nodes
}
