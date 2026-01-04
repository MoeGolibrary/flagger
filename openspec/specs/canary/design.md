# Flagger Architecture Design

## Context

Flagger is a progressive delivery operator for Kubernetes that automates the promotion of application deployments using
canary releases, A/B testing, and blue/green deployments. It integrates with various service mesh and ingress providers
to gradually shift traffic while analyzing key metrics.

## Goals / Non-Goals

- Goals:
    - Automate progressive delivery with minimal human intervention
    - Support multiple service mesh and ingress providers
    - Provide reliable metrics analysis for safe deployments
    - Maintain backward compatibility with existing configurations

- Non-Goals:
    - Replace existing CI/CD pipelines
    - Manage source code or build processes
    - Provide application monitoring beyond deployment metrics

## Decisions

- Decision: Use Kubernetes CRDs for declarative configuration
    - Reason: Aligns with Kubernetes patterns and allows for easy integration
    - Alternative considered: Configuration files in ConfigMap
    - Chosen because: Provides better validation and kubectl integration

- Decision: Implement provider-specific router implementations
    - Reason: Different service meshes and ingress controllers have unique APIs
    - Alternative considered: Generic traffic management API
    - Chosen because: Allows for optimal use of each provider's capabilities

- Decision: Use controller-runtime pattern with work queues
    - Reason: Follows Kubernetes operator best practices
    - Alternative considered: Polling-based approach
    - Chosen because: More efficient and responsive to changes

## Risks / Trade-offs

- Risk: Complex integration with multiple providers increases maintenance overhead
    - Mitigation: Abstract common functionality and maintain provider-specific implementations separately

- Risk: Traffic shifting might impact application performance during analysis
    - Mitigation: Allow for configurable analysis parameters and gradual traffic increase

## Migration Plan

- Phase 1: Implement core canary functionality for primary providers
- Phase 2: Add support for additional service mesh and ingress providers
- Phase 3: Enhance metrics analysis and alerting capabilities

## Open Questions

- How to handle multi-cluster deployments?
- What are the performance implications of monitoring many canaries simultaneously?