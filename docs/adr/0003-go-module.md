# ADR-0003: Single Go Module

## Status

Accepted

## Context

Nimbus is designed as a modular monolith consisting of multiple business domains within a single deployable application.

Go provides two approaches:

- Single module
- Workspace (go.work)

## Decision

Nimbus will use a single Go module.

## Rationale

- Simpler dependency management.
- Easier onboarding.
- Suitable for a modular monolith.
- Avoids unnecessary complexity.
- Easy to migrate later if services are extracted.

## Consequences

Positive

- Simpler project.
- Faster builds.
- Easier CI.

Negative

- Future service extraction may require module restructuring.
