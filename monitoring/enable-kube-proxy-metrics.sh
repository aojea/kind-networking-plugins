#!/bin/sh

# Full credit BenTheElder
# https://github.com/kubernetes-sigs/kind/blob/b6bc112522651d98c81823df56b7afa511459a3b/hack/ci/e2e-k8s.sh#L190-L205

# Get the current config
original_kube_proxy=$(kubectl get -oyaml -n=kube-system configmap/kube-proxy)
echo "Original CoreDNS config:"
echo "${original_kube_proxy}"
# Patch it
fixed_kube_proxy=$(
    printf '%s' "${original_kube_proxy}" | sed \
        's/\(.*metricsBindAddress:\)\( .*\)/\1 "0.0.0.0:10249"/' \
    )
echo "Patched kube-proxy config:"
echo "${fixed_kube_proxy}"
printf '%s' "${fixed_kube_proxy}" | kubectl apply -f -
# restart kube-proxy
kubectl -n kube-system rollout restart ds kube-proxy
