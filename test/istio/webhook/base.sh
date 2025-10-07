#!/usr/bin/env bash

# Base script for webhook E2E tests

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)

# Function to initialize test workloads
initialize_test_workloads() {
    echo '>>> Initializing test workloads'
    "$REPO_ROOT"/test/workloads/init.sh
}

# Function to create a canary
create_canary() {
    local canary_spec=$1
    echo '>>> Creating canary'
    echo "$canary_spec" | kubectl apply -f -
}

# Function to wait for canary initialization
wait_for_initialized() {
    echo '>>> Waiting for canary to be initialized'
    retries=20
    count=0
    ok=false
    until ${ok}; do
        kubectl -n test get canary/podinfo | grep 'Initialized' && ok=true || ok=false
        sleep 5
        count=$(($count + 1))
        if [[ ${count} -eq ${retries} ]]; then
            kubectl -n istio-system logs deployment/flagger
            echo "No more retries left"
            exit 1
        fi
    done
}

# Function to trigger canary deployment
trigger_deployment() {
    local image=$1
    echo '>>> Triggering canary deployment'
    kubectl -n test set image deployment/podinfo podinfod=${image}
}

# Function to wait for canary phase
wait_for_phase() {
    local phase=$1
    echo ">>> Waiting for canary to reach phase: ${phase}"
    retries=20
    count=0
    ok=false
    until ${ok}; do
        kubectl -n test get canary/podinfo | grep "${phase}" && ok=true || ok=false
        sleep 5
        count=$(($count + 1))
        if [[ ${count} -eq ${retries} ]]; then
            kubectl -n istio-system logs deployment/flagger
            echo "No more retries left"
            exit 1
        fi
    done
}

# Function to wait for canary completion
wait_for_completion() {
    echo '>>> Waiting for canary to complete'
    retries=30
    count=0
    ok=false
    until ${ok}; do
        kubectl -n test get canary/podinfo | grep 'Succeeded\|Failed' && ok=true || ok=false
        sleep 5
        count=$(($count + 1))
        if [[ ${count} -eq ${retries} ]]; then
            kubectl -n test describe deployment/podinfo
            kubectl -n test describe deployment/podinfo-primary
            kubectl -n istio-system logs deployment/flagger
            echo "No more retries left"
            exit 1
        fi
    done
}

# Function to verify canary success
verify_canary_success() {
    echo '>>> Verifying canary success'
    kubectl -n test get canary/podinfo | grep 'Succeeded'
    echo '>>> Canary completed successfully'
}

# Function to verify canary failure
verify_canary_failure() {
    echo '>>> Verifying canary failure'
    kubectl -n test get canary/podinfo | grep 'Failed'
    echo '>>> Canary failed as expected'
}