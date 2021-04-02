# Baremetal plugin

The Baremetal plugin extend KIND configuration to add additional networks to the nodes
in the cluster.

## Usage

```
./baremetal 
Simulate baremetal deployments using KIND clusters

Usage:
  baremetal [command]

Available Commands:
  create      Create a baremetal cluster
  delete      Delete the baremetal cluster
```

### Create


The plugin requires a configuration file like this:

```yaml
cluster:
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
  - role: worker
networks:
- storage
- external
```


That will create a cluster with 2 nodes, and each node will be attached to the specified networks: storage, external, in addition to the normal cluster network, that is named
after the cluster name and prefix with "bm-"

```
./baremetal create --config config.yaml 
Creating cluster "kind" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
```

Each network is an independent docker network, to avoid pullution the environment.

```
docker network ls
NETWORK ID     NAME       DRIVER    SCOPE
5b6b5f83995a   bm-kind    bridge    local
386420ca628f   external   bridge    local
012422695d18   storage    bridge    local
```

### Delete

Delete removes all the resources created.

```
./baremetal delete --config config.yaml 
```
