# Multizone plugin

The Multizone plugin allows to simulate a Kubernetes cluster deployed across multiple availability zones.

## Usage

```
 ./multizone -h
Simulate multizone deployments using KIND clusters

Usage:
  multizone [command]

Available Commands:
  create      Create a multizone cluster
  delete      Delete the multizone cluster
```

### Create


The plugin uses flags to define the number of zones and nodes per zone:

TODO: It deploys one control plane in the cluster. Best practices require
one control-plane per zone, it will nice also to simulate latency between zones.
See multicluster plugin for latency simulation.

```sh
 ./multizone create --nodes-zone 1 --zones 2
Creating cluster "kind" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (aojea/kindnode:1.22rc) 🖼
 ✓ Preparing nodes 📦 📦 📦  
```


That will create a cluster with 2 nodes, and each node will be placed in a different
availability zone. The zones are defined by the label `topology.kubernetes.io/zone`

```
kubectl get nodes --show-labels
NAME                 STATUS   ROLES                  AGE     VERSION   LABELS
kind-control-plane   Ready    control-plane,master   2m44s   v1.20.2   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=kind-control-plane,kubernetes.io/os=linux,node-role.kubernetes.io/control-plane=,node-role.kubernetes.io/master=
kind-worker          Ready    <none>                 2m10s   v1.20.2   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=kind-worker,kubernetes.io/os=linux,topology.kubernetes.io/zone=zone0
kind-worker2         Ready    <none>                 2m10s   v1.20.2   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=kind-worker2,kubernetes.io/os=linux,topology.kubernetes.io/zone=zone1
```


### Delete

Delete removes all the resources created.

```
./multizone delete
```
