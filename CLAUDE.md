# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

<!-- TODO: Adapt this section after creating a new project -->

myproject is a [TODO: Describe your project type, e.g., "multi-organisation SaaS application", "API backend", etc.] consisting of:

- **Backend**: Go/GraphQL API server with PostgreSQL database

[TODO: Add other components if applicable, e.g.:]
<!-- - **Frontend**: React web application -->
<!-- - **App**: React Native mobile application -->

[TODO: Briefly describe what the system manages/does]

## Tech Stack

- **Backend**: Go 1.25, GraphQL (gqlgen), PostgreSQL 16
- **Development**: Devbox for environment management, Process Compose for services

## Common Development Commands

The user needs to start the development environment using `devbox services up`. It is not your job to run services.

For backend-specific commands and patterns, see `backend/CLAUDE.md`.

## High-Level Architecture

The system follows a multi-tier architecture with clear separation of concerns:

- **Backend**: Domain-Driven Design with CQRS pattern, GraphQL API

### Key Domain Concepts

<!-- TODO: Adapt this section to describe your domain concepts -->

[TODO: List and describe your key domain entities, e.g.:]
<!-- - **Organisation**: Top-level tenant with settings -->
<!-- - **User**: Application users with roles and permissions -->
<!-- - **[Entity]**: Description of what this entity represents -->

### Authentication

- Accounts stored in backend database
- JWT tokens in secure HTTP-only cookies (session scoped to account)
- CSRF token cookie for mutation protection

## Testing Approach

- Backend: Standard Go testing with testify assertions and table-driven tests

## Code Quality

Always run linting before committing:

```bash
devbox run backend:lint
```
