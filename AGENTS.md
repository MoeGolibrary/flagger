# Flagger Development Guide

## Project Overview

Flagger is a progressive delivery tool for Kubernetes that automates canary deployments, A/B testing, and blue/green deployments using service mesh or ingress controllers. It integrates with various platforms like Istio, Linkerd, App Mesh, and others.

## Project Structure

```
.
├── artifacts          # Example configurations and deployment manifests
├── charts             # Helm charts for deploying Flagger and related components
├── cmd                # Main application entry points
│   ├── flagger        # Main Flagger controller
│   └── loadtester     # Load testing tool
├── hack               # Development scripts for code generation
├── kustomize          # Kustomize overlays for different service mesh setups
├── pkg                # Main source code
│   ├── apis           # API definitions for various service meshes
│   ├── canary         # Core canary deployment logic
│   ├── controller     # Main controller logic
│   ├── loadtester     # Load testing functionality
│   ├── router         # Routing implementations for different service meshes
│   └── ...            # Other packages (logger, metrics, notifier, etc.)
└── test               # Integration tests for different platforms
```

## Dev environment tips
- Use `make build` to build the Flagger binary for your platform
- Use `make test` to run all unit tests
- Use `make test-coverage` to run tests with coverage report
- Use `make fmt` to format Go code
- Use `make vet` for static analysis
- Use `make build-image` to build Docker image
- Use `make loadtester-build` to build the load tester image
- For development, run Flagger locally with `go run cmd/flagger/main.go` after setting up kubeconfig
- Use `make codegen` to regenerate client code after modifying CRDs

## Building and Running

### Building the Project

To build the Flagger binary:
```bash
make build
```

To build the Docker image:
```bash
make build-image TAG=<your-tag>
```

### Code Generation

Flagger uses Kubernetes code generation for its custom resources. After modifying CRDs, run:
```bash
make codegen
```

To verify code generation is up-to-date:
```bash
make test-codegen
```

## Testing instructions
- Find the CI plan in the .github/workflows folder
- Run `make test` to execute all tests
- Run `go test ./pkg/...` to run tests for all packages
- To focus on one package, run `go test ./pkg/controller/` (for example)
- Use `go test -v` for verbose output
- Fix any test or type errors until the whole suite is green
- After making changes, run `make fmt test-codegen` to ensure code is properly formatted
- Add or update tests for the code you change, even if nobody asked
- Use `make test-coverage` to run tests with coverage report

## E2E Testing with Istio

Flagger includes comprehensive end-to-end tests for Istio integration. These tests validate various deployment strategies including canary releases, blue/green deployments, and A/B testing.

### Test Structure

The Istio E2E tests are located in [test/istio/](file:///Users/hanyunpeng/Projects/flagger/test/istio) and include:

1. **install.sh** - Installs Istio and Flagger in the test environment
2. **run.sh** - Main test runner that executes all test scenarios
3. **test-canary.sh** - Tests canary deployment, blue/green, and A/B testing scenarios
4. **test-skip-analysis.sh** - Tests deployments with analysis skipped
5. **test-delegation.sh** - Tests virtual service delegation feature

### Running Istio E2E Tests

To run the Istio E2E tests:

1. Ensure you have Kubernetes Kind installed
2. Execute the test runner:
   ```bash
   cd test/istio
   ./run.sh
   ```

This will:
- Install Istio and Flagger
- Initialize test workloads
- Run canary deployment tests
- Run skip-analysis tests
- Run delegation tests

### Test Flow

The Istio E2E tests follow this workflow:

1. Install Istio and Flagger in the cluster
2. Create test namespace with istio-injection enabled
3. Deploy the load tester and podinfo test application
4. Test various deployment scenarios:
   - Canary deployments with traffic shifting
   - Blue/Green deployments
   - A/B testing with header-based routing
   - Skip analysis deployments
   - Virtual service delegation

### Key Test Components

1. **Metric Templates** - Custom Prometheus queries for measuring service performance
2. **Canary Objects** - Flagger's custom resources that define deployment strategies
3. **Webhooks** - Integration with load tester for acceptance and load testing
4. **Service Mesh Configuration** - Istio VirtualServices and DestinationRules managed by Flagger

## Key Components

1. **Core controller** - manages the deployment workflow in [pkg/controller/](file:///Users/hanyunpeng/Projects/flagger/pkg/controller)
2. **Canary CRD** - defines how deployments should be handled in [artifacts/flagger/crd.yaml](file:///Users/hanyunpeng/Projects/flagger/artifacts/flagger/crd.yaml)
3. **Router implementations** - for various service meshes and ingress controllers in [pkg/router/](file:///Users/hanyunpeng/Projects/flagger/pkg/router)
4. **Metrics integration** - with Prometheus for analysis in [pkg/metrics/](file:///Users/hanyunpeng/Projects/flagger/pkg/metrics)
5. **Notification system** - for Slack, MS Teams, etc. in [pkg/notifier/](file:///Users/hanyunpeng/Projects/flagger/pkg/notifier)

## Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run tests to ensure nothing is broken with `make test`
6. Commit with a signed-off message following the DCO
7. Push and create a pull request

## PR instructions
- Title format: [<component>] <Title> (e.g. [controller] Add manual step feature)
- Always run `make fmt test-codegen` and `make test` before committing
- Update documentation in README.md or /docs if needed
- Include unit tests for new functionality
- Follow existing code style and conventions
- Ensure all CI checks pass before merging

## Key Development Areas

1. **Adding support for new service meshes** - Implement new routers in [pkg/router/](file:///Users/hanyunpeng/Projects/flagger/pkg/router)
2. **Adding new metrics providers** - Extend the metrics functionality
3. **Enhancing notification providers** - Add new notification channels in [pkg/notifier/](file:///Users/hanyunpeng/Projects/flagger/pkg/notifier)
4. **Improving canary analysis** - Modify the analysis logic in [pkg/controller/](file:///Users/hanyunpeng/Projects/flagger/pkg/controller)