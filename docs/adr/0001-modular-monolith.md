# ADR-0001: Use a Modular Monolith Architecture

## Status

Accepted

## Date

2026-07-07

## Context

Nimbus is intended to be a production-grade backend platform that demonstrates professional software engineering practices while remaining easy to understand, develop, and deploy.

One of the earliest architectural decisions is whether to build the system as:

- A modular monolith
- A microservices architecture

Although microservices are widely used in large-scale systems, they introduce additional complexity such as:

- Service discovery
- Network communication
- Distributed transactions
- Independent deployments
- Distributed logging and tracing
- Infrastructure overhead

At the current stage of the project, these complexities do not provide sufficient value.

## Decision

Nimbus will be implemented as a **Modular Monolith**.

Business capabilities such as Authentication, Wallet, Ledger, Transfer, and Notification will each exist as independent modules within a single deployable application.

Each module will maintain clear boundaries and communicate through interfaces rather than tightly coupling implementation details.

## Rationale

A modular monolith provides several advantages:

- Simpler local development.
- Easier debugging.
- Single deployment artifact.
- Shared database transactions.
- Faster development iteration.
- Lower operational complexity.
- Easier onboarding for new contributors.

By enforcing module boundaries from the beginning, individual modules can later be extracted into independent services if future requirements justify that decision.

## Alternatives Considered

### Microservices

**Pros**

- Independent deployments
- Independent scaling
- Technology flexibility

**Cons**

- Increased operational complexity
- Network latency
- More difficult local development
- Distributed transactions
- Additional infrastructure requirements

Given the current goals of the project, these costs outweigh the benefits.

## Consequences

### Positive

- Faster feature development.
- Easier testing.
- Simpler deployments.
- Easier debugging.
- Cleaner developer experience.

### Negative

- Entire application is deployed as one unit.
- Scaling occurs at the application level rather than per module.
- Future extraction into microservices will require additional work.

## Future Considerations

If the project grows substantially or specific modules require independent scaling, modules such as Ledger or Notification may be extracted into standalone services while preserving their existing interfaces.
