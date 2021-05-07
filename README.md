# kind-networking-plugins

Plugins to extend KIND networking capabilities with plugins using the KIND API

These plugins were used for the Kubecon EU 2021 presentation
[Kubernetes Advanced Networking Testing with KIND](./Kubernetes _Advanced_Networking_Testing _with_KIND_Antonio_Ojea_Kubecon_2021.pdf)


## Overview

[KIND](https://kind.sigs.k8s.io/) is a tool for running local Kubernetes clusters using Docker container “nodes”.

KIND was primarily designed for testing Kubernetes itself, but may be used for local development or CI. This requires a strong focus in stability and resilience, thus adding new features is complicated. However, KIND exposes an API that can be leveraged for automation.

In the other hand, testing networking is always complicated, because it requires more complex
scenarios to be able to cover all the features. Traditionally, this was difficult to automate, but nowadays, current virtualization techniques, like containers and virual networks
make it possible.

This repository contains some example plugins to demonstrate how to extend KIND and automate complex Kubernetes clusters.

### Multicluster

The multicluster plugin allows to simulate multicluster environments, deploying independent clusters, and emulating the WAN between those cluster, so the user can define the bandwith, latency and the error rate.

[Usage](./multicluster/README.md)

### Multizone

The multizone plugin simulates a Cluster deployed across multiple availability zones.

[Usage](./multizone/README.md)


### Baremetal

The baremetal plugins demonstrates how to extend current KIND configuration to create new networks in the cluster.

[Usage](./baremetal/README.md)
