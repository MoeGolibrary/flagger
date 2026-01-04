# Flagger Controller Specification

## ADDED Requirements

### Requirement: Kubernetes Controller Pattern

The system SHALL implement a Kubernetes controller that watches for Canary custom resources and reconciles the desired
state by creating, updating, or deleting related Kubernetes objects.

#### Scenario: Canary resource creation

- **WHEN** a user creates a Canary custom resource
- **THEN** the controller enqueues the resource for processing
- **AND** the controller validates the Canary specification
- **AND** the controller creates necessary Kubernetes objects for the canary deployment

#### Scenario: Canary resource update

- **WHEN** a user updates an existing Canary resource
- **THEN** the controller detects the spec changes
- **AND** the controller reconciles the changes in the related Kubernetes objects
- **AND** the controller continues the canary analysis if already in progress

### Requirement: Work Queue Management

The system SHALL use a rate-limited work queue to process canary resources with proper error handling and retry
mechanisms.

#### Scenario: Successful resource processing

- **WHEN** the controller processes a canary resource successfully
- **THEN** the resource is removed from the work queue
- **AND** the controller continues processing other resources

#### Scenario: Resource processing error

- **WHEN** the controller encounters an error processing a canary resource
- **THEN** the resource is re-queued with rate limiting
- **AND** the controller logs the error appropriately
- **AND** the resource remains in the queue for retry

### Requirement: Event Broadcasting

The system SHALL broadcast Kubernetes events for important canary lifecycle events to enable monitoring and debugging.

#### Scenario: Canary event recording

- **WHEN** an important event occurs during canary processing
- **THEN** the controller records the event in the Kubernetes API
- **AND** the event is associated with the appropriate Canary resource
- **AND** the event contains relevant context and metadata

### Requirement: Finalizer Management

The system SHALL manage finalizers on Canary resources to ensure proper cleanup when resources are deleted.

#### Scenario: Canary deletion with revertOnDeletion

- **WHEN** a user deletes a Canary resource that has revertOnDeletion enabled
- **THEN** the controller finalizes the canary by reverting to stable state
- **AND** the controller removes finalizers after successful finalization
- **AND** the resource is removed from the API server

#### Scenario: Canary deletion without revertOnDeletion

- **WHEN** a user deletes a Canary resource that does not have revertOnDeletion enabled
- **THEN** the controller does not block the deletion
- **AND** the controller may remove any created resources during the normal cleanup process