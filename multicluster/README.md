# Multicluster plugin

The Multicluster plugin allows to simulate multiple Kubernetes cluster deployed across internet, and emulate latency, packet drops and bandwith constraints.

## Usage

```
 ./multicluster -h
Simulate multicluster deployments using KIND clusters

Usage:
  multicluster [command]

Available Commands:
  create      Create a multicluster cluster
  delete      Delete the multicluster cluster
```

### Create


The plugin define the cluster topology using a configuration file like this:

```yaml
clusters:
  cluster-us: 
    nodes: 2
    nodeSubnet: "172.88.0.0/16"
    podSubnet: "10.196.0.0/16"
    serviceSubnet: "10.96.0.0/16"
  cluster-eu:
    nodes: 2
    nodeSubnet: "172.89.0.0/16"
    podSubnet: "10.197.0.0/16"
    serviceSubnet: "10.97.0.0/16"
```

You can create a multicluster deployment:

```sh
 ./multizone create --config config.yaml
Creating cluster "kind" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (aojea/kindnode:1.22rc) 🖼
 ✓ Preparing nodes 📦 📦 📦  
```


That will create 2 KIND clusters in its independent networks. It will also create a container
that will act as gateway for those networks. This container is a WAN emulator, that will
allow the user to specify the latency, packet drops and bandwith for each fo the clusters.


```
docker ps
CONTAINER ID   IMAGE                        COMMAND                  CREATED         STATUS         PORTS                       NAMES
330fcede9fb3   kindest/node:v1.20.2         "/usr/local/bin/entr…"   3 seconds ago   Up 1 second    127.0.0.1:39959->6443/tcp   cluster-eu-control-plane
5db3e1cd057a   kindest/node:v1.20.2         "/usr/local/bin/entr…"   3 seconds ago   Up 1 second                                cluster-eu-worker
0279df468048   quay.io/aojea/wanem:latest   "sleep infinity"         4 seconds ago   Up 4 seconds                               wan-kind
```

### Delete

Delete removes all the resources created.

```
./multicluser delete --config config.yml
```
