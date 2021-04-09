#!/bin/sh

# get prometheus URL from the Service (servicePort = 8080)
CLUSTER_IP=$(kubectl get svc prometheus-service -n monitoring -o jsonpath="{.spec.clusterIP}")
# create a snapshot of the db
docker exec -it kind-control-plane curl -XPOST http://${CLUSTER_IP}:8080/api/v1/admin/tsdb/snapshot
# get the prometheus database
POD_NAME=$(kubectl -n monitoring get pods -o jsonpath='{.items[0].metadata.name}')
kubectl cp monitoring/${POD_NAME}:/prometheus/snapshots /tmp/snapshots
# or a tarball
# kubectl -n monitoring exec $POD_NAME -- tar cvf - /prometheus/snapshots > snapshot.tar

# to see the metrics locally
# create a fake prometheus config so it does not complain

# touch prometheus.yaml

# init the container mounting the folder with the metrics

# docker run --rm -p 9090:9090 -uroot -v /tmp/snapshots:/prometheus prom/prometheus --config prometheus.yml --storage.tsdb.path=/prometheus

# bonus using grafana (admin:admin) and add prometheus as source
# docker run -d -p 3000:3000 grafana/grafana
