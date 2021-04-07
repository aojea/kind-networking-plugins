# monitoring

Modify kube-proxy to expose the metrics server in all interfaces, so we can
scrape them:

```
./enable-kube-proxy-metrics.sh
```

Install a prometheus instance that obtains metrics from the different components
of the cluster.

```
kubectl apply -f monitoring.yaml
```
It creates a NodePort service so you can use the prometheus instance in one of
the nodes IPs.

Ref:
- https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
