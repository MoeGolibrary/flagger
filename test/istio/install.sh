#!/usr/bin/env bash

set -o errexit

ISTIO_VER="1.27.1"
REPO_ROOT=$(git rev-parse --show-toplevel)

echo ">>> Checking if Istio ${ISTIO_VER} is already installed"
if kubectl get deployment istiod -n istio-system >/dev/null 2>&1; then
    INSTALLED_VERSION=$(kubectl get deployment istiod -n istio-system -o jsonpath='{.spec.template.spec.containers[0].image}' | cut -d ':' -f 2)
    if [ "$INSTALLED_VERSION" = "$ISTIO_VER" ]; then
        echo ">>> Istio ${ISTIO_VER} is already installed"
        # Just install prometheus and flagger
        kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/prometheus.yaml 2>/dev/null || true
        kubectl -n istio-system rollout status deployment/prometheus || true
        
        echo '>>> Installing Flagger'
        kubectl apply -k ${REPO_ROOT}/kustomize/istio
        
        kubectl -n istio-system set image deployment/flagger flagger=pingxin/flagger:latest
        kubectl -n istio-system rollout status deployment/flagger
        exit 0
    else
        echo ">>> Different Istio version ($INSTALLED_VERSION) detected, reinstalling with ${ISTIO_VER}"
    fi
fi

mkdir -p ${REPO_ROOT}/bin

echo ">>> Downloading Istio ${ISTIO_VER}"
cd ${REPO_ROOT}/bin && \
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VER} sh -

echo ">>> Installing Istio ${ISTIO_VER}"
${REPO_ROOT}/bin/istio-${ISTIO_VER}/bin/istioctl manifest install --set profile=default --skip-confirmation \
  --set values.pilot.resources.requests.cpu=100m \
  --set values.pilot.resources.requests.memory=100Mi

kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/prometheus.yaml
kubectl -n istio-system rollout status deployment/prometheus

kubectl -n istio-system get all

echo '>>> Installing Flagger'
kubectl apply -k ${REPO_ROOT}/kustomize/istio

kubectl -n istio-system set image deployment/flagger flagger=pingxin/flagger:latest
kubectl -n istio-system rollout status deployment/flagger

echo ">>> Istio ${ISTIO_VER} and Flagger installed successfully"