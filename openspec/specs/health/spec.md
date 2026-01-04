# Flagger Health Checking Specification

## ADDED Requirements

### Requirement: Deployment Health Monitoring

The system SHALL continuously monitor the health of primary and canary deployments during the analysis process.

#### Scenario: Deployment readiness check

- **WHEN** the system performs health checks on deployments
- **THEN** the system verifies that all pods are in ready state
- **AND** the system checks that deployments have the expected number of replicas
- **AND** the system updates the canary status with health information

#### Scenario: Health check failure

- **WHEN** a deployment fails health checks
- **THEN** the system records the health failure in the canary status
- **AND** the system may pause or rollback the canary analysis based on configuration
- **AND** the system sends appropriate notifications about the health issue

### Requirement: Service Health Validation

The system SHALL validate the health of services exposed by canary deployments.

#### Scenario: Service connectivity check

- **WHEN** the system validates service health
- **THEN** the system verifies that services are responding to requests
- **AND** the system checks service availability through configured endpoints
- **AND** the system records service health status in the canary analysis

### Requirement: Health-Based Decision Making

The system SHALL use health information to make decisions about canary progression.

#### Scenario: Health-based rollback

- **WHEN** deployment health deteriorates during canary analysis
- **THEN** the system evaluates whether to continue or rollback
- **AND** if health thresholds are exceeded, the system initiates rollback
- **AND** the system updates the canary status with the decision and reason

#### Scenario: Health-based pause

- **WHEN** deployment health is questionable but not critical
- **THEN** the system may pause the canary analysis
- **AND** the system waits for health to improve or for manual intervention
- **AND** the system continues to monitor health during the pause