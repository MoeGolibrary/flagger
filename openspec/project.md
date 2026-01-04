# Flagger Project Context

## Purpose

Flagger is a progressive delivery tool that automates the release process for applications running on Kubernetes. It
reduces the risk of introducing a new software version in production by gradually shifting traffic to the new version
while measuring metrics and running conformance tests. Flagger implements several deployment strategies (Canary
releases, A/B testing, Blue/Green mirroring) and integrates with various Kubernetes ingress controllers, service mesh,
and monitoring solutions.

## Tech Stack

- Go (Golang) - Primary backend language
- Kubernetes - Container orchestration platform
- Controller-runtime - Kubernetes controller framework
- Prometheus - Metrics collection and monitoring
- Helm - Package management for Kubernetes
- Docker - Containerization
- Various service mesh technologies (Istio, Linkerd, App Mesh, Kuma, OSM)

## Project Conventions

### Code Style

- Go code follows standard Go formatting (gofmt)
- Use descriptive variable and function names
- Follow Kubernetes controller patterns and best practices
- Write comprehensive unit tests for all business logic
- Use zap for logging with structured logging approach
- Follow Kubernetes naming conventions for resources and labels

### Architecture Patterns

- Kubernetes controller pattern with custom resource definitions (CRDs)
- Event-driven architecture using informers and work queues
- Factory pattern for creating different implementations based on provider
- Interface segregation for different components (router, canary, tracker)
- Separation of concerns between controller, business logic, and infrastructure

### Testing Strategy

- Unit tests for all core business logic functions
- Integration tests for Kubernetes API interactions
- End-to-end tests for complete canary deployment scenarios
- Use test fixtures for consistent test data
- Mock external dependencies in unit tests

### Git Workflow

- Feature branches for new functionality
- Squash and merge for pull requests
- Semantic commit messages following conventional commits
- Branch names should be descriptive and follow kebab-case format

## Domain Context

Flagger operates in the progressive delivery domain, managing canary deployments, A/B tests, and blue/green deployments.
It integrates with various service mesh and ingress providers to gradually shift traffic from stable to new application
versions based on defined metrics and success criteria. The system manages the full lifecycle of canary deployments
including initialization, traffic shifting, metric analysis, promotion, and rollback.

## Important Constraints

- Must maintain backward compatibility with existing Canary CRDs
- Should not modify user's existing deployments directly
- Must handle multi-tenant environments safely
- Should gracefully handle network partitions and API server outages
- Must support both Kubernetes-native and service mesh traffic management

## External Dependencies

- Kubernetes API server for resource management
- Prometheus for metrics collection and analysis
- Various ingress controllers and service mesh providers
- External monitoring and alerting systems
- Container registries for image management