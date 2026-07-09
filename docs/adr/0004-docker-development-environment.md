# ADR-0004: Containerized Local Development Environment

## Status

Accepted

## Date

2026-07-07

## Context

Nimbus depends on external services such as PostgreSQL and Redis. Requiring contributors to install and configure these services manually can lead to inconsistent environments and onboarding difficulties.

## Decision

Nimbus will use Docker Compose to provision local development dependencies.

The application itself will continue to run on the host machine during development, while infrastructure services such as PostgreSQL and Redis will run in containers.

## Rationale

- Consistent development environments
- Faster onboarding
- Easy cleanup and recreation
- Reduced dependency on host operating system configuration

## Alternatives Considered

### Install services directly on the host

**Pros**

- Slightly lower resource usage

**Cons**

- OS-specific setup
- Version inconsistencies
- Harder onboarding

## Consequences

### Positive

- Predictable development environment
- Easier collaboration
- Simplified setup documentation

### Negative

- Docker Desktop/Engine becomes a prerequisite
- Slightly higher memory usage
