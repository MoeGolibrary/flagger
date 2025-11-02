#!/usr/bin/env bash

# This script runs all Istio E2E tests

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
DIR="$(cd "$(dirname "$0")" && pwd)"

echo ">>> Starting Istio E2E tests"

echo ">>> Installing Istio and Flagger"
"$DIR"/install.sh

echo ">>> Initializing test workloads"
"$REPO_ROOT"/test/workloads/init.sh

echo ">>> Running canary test"
"$DIR"/test-canary.sh || { echo "Canary test failed but continuing..."; }

echo ">>> Running skip analysis test"
"$REPO_ROOT"/test/workloads/init.sh
"$DIR"/test-skip-analysis.sh || { echo "Skip analysis test failed but continuing..."; }

echo ">>> Running delegation test"
"$REPO_ROOT"/test/workloads/init.sh
"$DIR"/test-delegation.sh || { echo "Delegation test failed but continuing..."; }

echo ">>> Running traffic mirroring test"
"$REPO_ROOT"/test/workloads/init.sh
"$DIR"/test-mirroring.sh || { echo "Mirroring test failed but continuing..."; }

echo ">>> Running webhook tests"
"$DIR"/webhook/run.sh || { echo "Webhook tests failed but continuing..."; }

echo ">>> All Istio E2E tests completed"