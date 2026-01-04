# Flagger Canary Specification

## ADDED Requirements

### Requirement: Canary Deployment Management

The system SHALL manage canary deployments by creating and orchestrating a series of Kubernetes objects including
deployments, services, and traffic routing rules to gradually shift traffic from stable to new application versions.

#### Scenario: Successful canary deployment creation

- **WHEN** a user creates a Canary CRD with valid configuration
- **THEN** the system creates necessary Kubernetes objects (primary and canary deployments, services, etc.)
- **AND** the system begins monitoring metrics and managing traffic distribution

#### Scenario: Invalid canary configuration

- **WHEN** a user creates a Canary CRD with invalid configuration
- **THEN** the system reports the validation error in the Canary status
- **AND** the system does not proceed with deployment creation

### Requirement: Traffic Shifting

The system SHALL gradually shift traffic from the stable version to the canary version based on the defined analysis
schedule and success criteria.

#### Scenario: Weight-based traffic shifting

- **WHEN** a canary analysis is in progress and metrics are within acceptable thresholds
- **THEN** the system gradually increases traffic to the canary version according to stepWeight configuration
- **AND** the system updates routing rules to reflect new traffic distribution

#### Scenario: Traffic rollback on failure

- **WHEN** a canary analysis detects metric failures exceeding threshold
- **THEN** the system immediately stops traffic shifting
- **AND** the system rolls back to the stable version
- **AND** the system updates the canary status to reflect the rollback

### Requirement: Metrics Analysis

The system SHALL continuously monitor and analyze application metrics during canary deployments to determine success or
failure.

#### Scenario: Prometheus metric evaluation

- **WHEN** the system evaluates Prometheus-based metrics during a canary analysis
- **THEN** the system queries Prometheus for defined metrics (success rate, request duration, etc.)
- **AND** the system compares metric values against defined thresholds
- **AND** the system records the analysis results in the canary status

#### Scenario: Custom metric evaluation

- **WHEN** the system evaluates custom metrics defined via MetricTemplate CRDs
- **THEN** the system executes the custom metric query
- **AND** the system compares results against defined thresholds
- **AND** the system updates the canary status accordingly

### Requirement: Webhook Integration

The system SHALL support executing webhooks at different stages of the canary deployment process for custom validation
and testing.

#### Scenario: Pre-rollout webhook execution

- **WHEN** a canary analysis reaches the pre-rollout stage
- **THEN** the system executes configured pre-rollout webhooks
- **AND** the system waits for webhook completion before proceeding
- **AND** if the webhook fails, the canary analysis is rolled back

#### Scenario: Rollout webhook execution

- **WHEN** a canary analysis is in progress
- **THEN** the system executes configured rollout webhooks (e.g., load testing)
- **AND** the system monitors webhook execution and reports status

### Requirement: Service Mesh Integration

The system SHALL integrate with multiple service mesh technologies to manage traffic routing and metrics collection.

#### Scenario: Istio integration

- **WHEN** a canary is configured with provider set to 'istio'
- **THEN** the system creates Istio VirtualService and DestinationRule resources
- **AND** the system manages traffic splitting using Istio's traffic management capabilities

#### Scenario: Linkerd integration

- **WHEN** a canary is configured with provider set to 'linkerd'
- **THEN** the system creates Linkerd ServiceProfile resources
- **AND** the system manages traffic splitting using Linkerd's traffic management capabilities

### Requirement: Alerting and Notifications

The system SHALL provide alerting capabilities to notify stakeholders about canary deployment status changes.

#### Scenario: Slack notification on canary events

- **WHEN** a canary deployment phase changes (progressing, succeeded, failed)
- **THEN** the system sends appropriate notifications to configured alert providers
- **AND** the system includes relevant context about the canary status

#### Scenario: Alert provider integration

- **WHEN** an AlertProvider CRD is configured
- **THEN** the system uses the configured provider to send notifications
- **AND** the system supports multiple alert provider types (Slack, Discord, etc.)