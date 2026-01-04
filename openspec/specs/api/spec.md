# Flagger API Specification

## ADDED Requirements

### Requirement: Kubernetes Custom Resource Definition (CRD) Support

The system SHALL provide and manage custom resource definitions for canary deployments and related functionality.

#### Scenario: Canary CRD management

- **WHEN** a user creates a Canary custom resource
- **THEN** the system validates the resource against the CRD schema
- **AND** the system processes the resource according to its specification
- **AND** the system updates the resource status with current state information

#### Scenario: MetricTemplate CRD support

- **WHEN** a user defines custom metrics using MetricTemplate CRDs
- **THEN** the system processes the template to execute custom metric queries
- **AND** the system validates the template syntax and parameters
- **AND** the system uses the template during canary analysis

### Requirement: Status Management

The system SHALL maintain and update status information for managed resources.

#### Scenario: Canary status updates

- **WHEN** the state of a canary changes during analysis
- **THEN** the system updates the status field of the Canary resource
- **AND** the system includes phase information (progressing, succeeded, failed, etc.)
- **AND** the system records relevant metrics and analysis results

#### Scenario: Status reconciliation

- **WHEN** the controller restarts or reconciles resources
- **THEN** the system ensures status information reflects the current state
- **AND** the system updates any stale status information
- **AND** the system continues canary analysis from the appropriate point

### Requirement: Configuration Validation

The system SHALL validate configuration parameters before applying changes.

#### Scenario: Configuration validation on update

- **WHEN** a user updates a Canary resource configuration
- **THEN** the system validates the new configuration parameters
- **AND** the system rejects invalid configurations with appropriate error messages
- **AND** the system continues with valid configurations while preserving ongoing analysis