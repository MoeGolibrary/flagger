#!/usr/bin/env bash

# Test for manual traffic control webhook
# Note: This test is a placeholder as the manual traffic control webhook functionality is not yet implemented in the loadtester

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control via webhook (Placeholder)'

echo '>>> This test is a placeholder because the manual traffic control webhook functionality is not yet implemented in the loadtester'
echo '>>> To implement this test, the loadtester would need to support the following endpoint:'
echo '>>>   - /traffic/ (accepting POST requests with JSON payload)'
echo '>>> '
echo '>>> The webhook would be configured as follows:'
echo '>>>'
echo '>>> webhooks:'
echo '>>>   - name: manual-traffic-control'
echo '>>>     type: manual-traffic-control'
echo '>>>     url: http://flagger-loadtester.test/traffic/'
echo '>>>'
echo '>>> To pause traffic, a request would be sent to:'
echo '>>>   curl -d '\'{"paused": true}\'' http://localhost:8080/traffic/'
echo '>>>'
echo '>>> To resume traffic, a request would be sent to:'
echo '>>>   curl -d '\'{"paused": false}\'' http://localhost:8080/traffic/'
echo '>>>'
echo '>>> To set a specific weight, a request would be sent to:'
echo '>>>   curl -d '\'{"weight": 30}\'' http://localhost:8080/traffic/'

echo '✔ Manual Traffic Control via webhook test documented'