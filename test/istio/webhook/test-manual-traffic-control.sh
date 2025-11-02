#!/usr/bin/env bash

# Test for manual traffic control webhook

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control via webhook'

# Initialize test workloads
initialize_test_workloads

# Create canary with manual traffic control webhook
echo '>>> Creating canary with manual traffic control webhook'
kubectl apply -f - <<EOF
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
  namespace: test
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podinfo
  service:
    port: 9898
    gateways:
    - istio-system/public-gateway
    hosts:
    - app.example.com
  analysis:
    interval: 10s
    threshold: 5
    maxWeight: 100
    stepWeight: 20
    webhooks:
    - name: manual-traffic-control
      type: manual-traffic-control
      url: http://flagger-loadtester.test/traffic/state
EOF

# Wait for canary to be initialized
wait_for_initialized

# Show initial state of canary
echo '>>> Initial Canary status:'
kubectl -n test get canary podinfo -o yaml

# Show initial VirtualService
echo '>>> Initial VirtualService:'
kubectl -n test get virtualservice podinfo -o yaml

# Trigger a deployment to start canary analysis
trigger_deployment stefanprodan/podinfo:3.1.1

# Wait for canary to start progressing
wait_for_phase Progressing

echo '>>> Waiting 30 seconds for canary to progress...'
sleep 30

# Show canary status during progression
echo '>>> Canary status during progression:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService during progression
echo '>>> VirtualService during progression:'
kubectl -n test get virtualservice podinfo -o yaml

# Send manual traffic control command
echo '>>> Send manual traffic control command to at 55% weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -H "Canary-Name: podinfo" -H "Canary-Namespace: test" -d '{"weight": 55, "paused": false }' http://flagger-loadtester.test/traffic/

# Show canary status after setting weight
echo '>>> Canary status after setting weight to 55%:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService after setting weight
echo '>>> VirtualService after setting weight to 55%:'
kubectl -n test get virtualservice podinfo -o yaml


# Wait a bit for the command to be processed
echo '>>> Waiting 30 seconds for command processing...'
sleep 30

# Show canary status after setting weight
echo '>>> Canary status after setting weight to 55%:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService after setting weight
echo '>>> VirtualService after setting weight to 55%:'
kubectl -n test get virtualservice podinfo -o yaml

# Send manual traffic control command
echo '>>> Send manual traffic control command to pause'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -H "Canary-Name: podinfo" -H "Canary-Namespace: test" -d '{"weight": 33, "paused": true }' http://flagger-loadtester.test/traffic/

# Show canary status after pausing
echo '>>> Canary status after pausing at 33%:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService after pausing
echo '>>> VirtualService after pausing at 33%:'
kubectl -n test get virtualservice podinfo -o yaml


# Wait a bit for the command to be processed
echo '>>> Waiting 30 seconds for pause command processing...'
sleep 30

# Show canary status after pausing
echo '>>> Canary status after pausing at 33%:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService after pausing
echo '>>> VirtualService after pausing at 33%:'
kubectl -n test get virtualservice podinfo -o yaml

echo '>>> Sending manual traffic control command to at 77% weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -H "Canary-Name: podinfo" -H "Canary-Namespace: test" -d '{"weight": 77, "paused": false }' http://flagger-loadtester.test/traffic/

# Show canary status after pausing
echo '>>> Canary status after pausing at 77%:'
kubectl -n test get canary podinfo -o yaml

# Show VirtualService after pausing
echo '>>> VirtualService after pausing at 77%:'
kubectl -n test get virtualservice podinfo -o yaml

# Wait a bit for the command to be processed
echo '>>> Waiting 30 seconds for pause command processing...'
sleep 30

# Wait for canary to complete
wait_for_completion

# Show final canary status
echo '>>> Final Canary status:'
kubectl -n test get canary podinfo -o yaml

# Show final VirtualService
echo '>>> Final VirtualService:'
kubectl -n test get virtualservice podinfo -o yaml

# Show Flagger logs
echo '>>> Flagger logs:'
kubectl -n istio-system logs deployment/flagger --tail=100

echo '✔ Manual Traffic Control via webhook test passed'