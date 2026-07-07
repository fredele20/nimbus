# ADR-0002: Adopt a Layered Architecture

## Status

Accepted

## Date

2026-07-07

## Context

Nimbus requires an application structure that separates responsibilities while remaining easy to understand for contributors.

Several architectural patterns were considered:

- MVC
- Layered Architecture
- Clean Architecture
- Hexagonal Architecture (Ports & Adapters)
- Onion Architecture

While Clean and Hexagonal architectures provide excellent separation of concerns, they often introduce additional abstractions that are unnecessary for a project of this size.

## Decision

Nimbus will adopt a **Layered Architecture** consisting of the following layers:

```
HTTP
↓
Handler
↓
Service
↓
Repository
↓
Database
```

Business logic will reside within the Service layer.

Repositories will be responsible only for persistence.

Handlers will only translate HTTP requests and responses.

## Layer Responsibilities

### Handler

Responsibilities:

- Parse requests
- Validate input
- Call services
- Format responses

Must NOT contain business logic.

### Service

Responsibilities:

- Business rules
- Transactions
- Authorization
- Coordination between repositories

Must remain independent of HTTP.

### Repository

Responsibilities:

- Database interaction
- Queries
- Persistence

Must not contain business logic.

### Database

Responsible for persistent storage using PostgreSQL.

## Rationale

Layered Architecture provides:

- Clear separation of concerns.
- Easier testing.
- Lower cognitive overhead.
- Familiarity for Go developers.
- Simpler onboarding.

It also allows future migration toward Clean Architecture if increased complexity justifies additional abstraction.

## Alternatives Considered

### MVC

Pros

- Simple
- Familiar

Cons

- Business logic often leaks into controllers.
- Harder to maintain as projects grow.

### Clean Architecture

Pros

- Strong separation of concerns.
- Framework independence.

Cons

- More interfaces.
- Higher initial complexity.
- Can become over-engineered for medium-sized projects.

### Hexagonal Architecture

Pros

- Excellent testability.
- Infrastructure independence.

Cons

- Additional abstraction.
- Steeper learning curve.

## Consequences

### Positive

- Predictable project structure.
- Easier maintenance.
- Easier unit testing.
- Better code organization.

### Negative

- Some business operations may require coordination across multiple services.
- Future architectural evolution may introduce additional abstraction layers.

## Future Considerations

As Nimbus grows, internal modules can adopt additional patterns such as domain services or ports-and-adapters without requiring a complete restructuring of the project.
