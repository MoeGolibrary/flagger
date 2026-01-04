# Flagger OpenSpec Documentation

This directory contains OpenSpec documentation for the Flagger progressive delivery system to help AI assistants and
developers understand the codebase.

## Structure

- [project.md](project.md) - Project context, tech stack, and conventions
- [specs/](specs/) - Feature specifications organized by component
    - [api/](specs/api/spec.md) - API and CRD specifications
    - [canary/](specs/canary/spec.md) - Core canary deployment functionality
    - [config-tracking/](specs/config-tracking/spec.md) - Configuration change tracking
    - [controller/](specs/controller/spec.md) - Kubernetes controller patterns
    - [deployment/](specs/deployment/spec.md) - Deployment management
    - [health/](specs/health/spec.md) - Health checking functionality
    - [loadtester/](specs/loadtester/spec.md) - Load testing and webhook execution
    - [metrics/](specs/metrics/spec.md) - Metrics collection and analysis
    - [notifier/](specs/notifier/spec.md) - Notification and alerting systems
    - [router/](specs/router/spec.md) - Traffic routing functionality
- [changes/](changes/) - Change proposals and implementations
- [proposal.md](proposal.md) - Current documentation proposal
- [tasks.md](tasks.md) - Implementation tasks

## Purpose

This documentation helps AI assistants understand:

- The architecture and design patterns used in Flagger
- The domain context of progressive delivery
- How different components interact
- The expected behavior of core features
- Development conventions and best practices