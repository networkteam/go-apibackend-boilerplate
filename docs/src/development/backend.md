# Backend

The backend is implemented in Go (without using a framework).

The code of the backend is located in the `backend` folder in the repository.

## General structure

- The GraphQL API is implemented using [gqlgen](https://gqlgen.com/).
- It uses a lightweight [CQRS](https://en.wikipedia.org/wiki/Command%E2%80%93query_separation) architecture,
  where _queries_ (reads) are separated from _commands_ (writes).
- The domain package is free of implementation details and encapsulates domain logic, validation, and common types.
- Each _command_ either succeeds (no error) or fails (returns an error) - it must be handled transactionally.

    !!! note

        Multiple commands using a single transaction are not yet considered and can usually be bypassed by using eventual consistency.
        It can also be solved by choosing another granularity of commands (e.g. bundling multiple commands into a single command that manages a transaction).

- Commands are handled by methods of `handler.Handler` - it validates the command, performs authorization checks and
  executes queries via the repository and can call external services or send emails.
- Queries are executed by methods of `finder.Finder` - it validates the query, performs authorization checks and
  executes the query via repository functions.
- Persistence is implemented via pure SQL (using pgx) with a dedicated code generator ([construct](https://github.com/networkteam/construct/)) for mapping fields.
- Queries are implemented as functions in `persistence/repository` using [qrb](https://github.com/networkteam/qrb) for building PostgreSQL queries.

    !!! info

        SQL SELECTS mostly return a single JSON value for direct mapping to Go structs. Aggregation and loading of references via subselects should be used to optimize the number of queries (and avoid the `N+1` select problem). The field selection from GraphQL is forwarded to query options, that can eagerly include the needed relations and aggregations.

- For authentication, JSON Web Tokens (JWT) are used in an HTTP-only cookie as well as CSRF tokens.
- Authorization is done either before applying a command or executing a query by checking permissions.
    - Authorization is built around simple functions that check commands or queries (all part of the `domain` package) based on an authentication context. If needed, additional data is passed to these functions.

### Backend architecture

#### Package structure

`api`

: `graph`

      :  Contains the GraphQL API with schemas (`*.graphqls`).
         The code for resolver methods and API models is generated via [gqlgen](https://gqlgen.com/).

         Some API Models are mapped to custom implementations in `gqlgen.yml` to map custom types.

         All resolvers generally require authentication, which is provided as `authentication.AuthContext` in `context.Context`.
         This is mapped via a GraphQL field middleware.
         The `@bypassAuthentication` directive can be used to disable this for individual resolvers so that they can be called publicly
         (e.g. registration and login).

         Query resolvers map the arguments for filtering and pagination to a query type from the `domain` package
         and call a finder method to perform the actual query - including authorization.

         Mutation resolvers map the input data to a command type from the `domain` package.
         A handler method is called with the command and returns an `error` if the operation was not
         successful. Beside general errors, which land as GraphQL error in the response, there are also domain-specific errors
         e.g. for validation or different error conditions. These are returned as `domain.FieldError` from the handler
         (potentially wrapped in another error). For mutations that return a `Result` GraphQL type
         these specific errors are mapped to the result.

         Functions for converting API types to domain types (and vice versa for input types) are stored in the `helper` package and
         are used in the resolvers. This way, differences between the external and internal models can be easily and explicitly
         compensated and reduce the coupling to the API.

         `schema.graphqls`

         : The GraphQL schema for the frontend.

         `authentication.graphqls`

         : The GraphQL schema for login / logout and login status retrieval.

         `admin.graphqls`

         : The GraphQL schema for the admin interface.

: `handler`

      :  HTTP handler for use in the server or in the tests. Besides the GraphQL handler for the API there can be
         additional handlers e.g. for reports or download of dynamic data.

: `http`

      :  HTTP middlewares for authentication, logging and Sentry binding are stored in `api/http/middleware` and are
         assembled in `http.MiddlewareStack` to re-use the full middleware stack for different handlers in an
         HTTP mux (router).

`cli`

:    Contains the main packages for various CLI programs. Outside of development only `cli/ctl` is relevant, which bundles all commands for running and managing the backend.

     The most important entry point here is `cli/ctl/cmd_server.go` - where the actual server is implemented.

`domain`

:    Contains domain models, custom types and functions for business logic. This package should not have any dependencies to other packages
and forms the innermost layer of the architecture.

     *Commands* are the input data for all writing operation of handlers and validate their own data.
     Based on CQRS, write operations do not take place on the model (e.g. entities or some form of active record),
     but via commands and handler methods. Identifiers are generated when creating a command that will create
     a new resource. That way the caller already knows the identifiers of resources that will be created in advance.

     *Queries* include identifiers or filters to load single or multiple resources. They are mostly plain structs
     used by finder methods.

`handler`

:    *handler* methods perform the logic of processing a single command. Each command type has a particular method.
     Commands do not return values: only an optional error is returned in case of an error.
     A handler method will validate the command and perform authorization checks based on the passed `context.Context`.
     Processing of commands must be transactional per command.

     *Jobs* are integration services which are executed via a cron package in the server process according to fixed time rules.
     A job is implemented like a handler without a command - but is itself responsible for logging / error handling.

`mail`

:    A `mail.Mailer` sends emails based on Go templates. It uses a `mail.Sender` that uses SMTP or a fixture
     implementations for tests.

     Mail messages are struct types that contain all the data that is needed to build the email.
     They implement `mail.MessageProvider` and create a `*gomail.Message` by using an embedded template.
     A message can send headers or attach files if applicable.

`persistence`

:  `migrations`

      :  All changes to the database schema are mapped via migrations for automatic execution during deployments,
         see [Create a new migration](#create-a-new-migration).

:  `repository`

      :  Repository functions for queries and write operations.
         Mappings are created via [construct](https://github.com/networkteam/construct) from struct tags of the models
         to reduce boilerplate code. Repositories do not contain logic but implement the reading of data as well as
         the change of data over change sets.

         Most of the time either identifiers for single results or query types and pagination information for
         multiple results are passed as arguments. The result can either be a model or a result type just
         for a special case of query (e.g. reports or aggregation) - but it should always live in the `domain`
         package to pass it freely through the layers without creating a direct coupling to the `repository` package.

`security`

:  `authentication`

      :  Data types and functions for authentication via auth token and CSRF Token.
         The authentication information is stored in `authentication.AuthContext` and passed to middlewares and handlers for use via `context.Context`.

         Authentication is based on [JWT](https://jwt.io) tokens with secrets bound to each account.
         These are transmitted as HTTP-only, secure cookies.
         Thus, a session store can be omitted and sessions are preserved through the client without storing state in the backend.
         By using account-specific secrets, sessions can still be effectively invalidated, e.g. after a password has been changed.

         A CSRF token is supplied by the client in the `X-CSRF-Token` header and protects against cross-site request forgery attacks.

:  `authorization`

      :  Authorization logic is based on domain types. The functions are called by query resolvers and finders,
         to check access to an operation. All the information for a decision is passed to the functions -
         these have no context other than the `AuthContext`.

`test`

:    Helper for tests and fixtures.

     Tests use fixed time values with `test.FixedTimeSource` for reproducible execution of time-based logics.

:  !!! warning

         Direct usage of `time.Now()` should be prevent throughout the whole application and
         `domain.TimeSource` should be used instead as a dependency (or passed as an argument).
         This makes time-related behaviour much easier to test.

#### Dependency management

No special techniques like containers are used for passing dependencies.
Important dependencies for API handlers and resolvers are collected in `api.ResolverDependencies`.
The CLI commands create dependencies (database, mailer, etc.) with the given command-line flags
and pass them to the corresponding constructor functions.

Since most dependencies are specified in the form of interfaces, they can be easily exchanged in tests or the
implementation can be changed by providing different CLI options.

!!! note

    It might seem that this is too simple, but it works great and reduces layers of hard to understand and follow
    code. It is very simple to find usages of dependencies (e.g. "Where is the SMTP sender for a mailer created?")
    and code paths can be followed just by following method calls.

## Setting up a local development environment

### Requirements

- **Go** (>=1.18)
- **PostgreSQL** (>=13)

### 1. Set up the database

```shell
cd backend
```

#### Create new database:

```shell
createdb myproject-dev
```

#### Execute migrations:

```shell
go run ./cli/ctl migrate up
```

#### Create and prepare database for tests:

```shell
createdb myproject-test
go run ./cli/ctl test preparedb
```

!!! info

    Why is it necessary to prepare the database?
    Tests run in parallel and PostgreSQL can have race conditions with `CREATE EXTENSION` on a single database.

### 2. Import fixtures

Fixed data (fixtures) can be imported into the database for development:

```shell
go run ./cli/ctl fixtures --confirm
```

!!! warning

    All existing data in the database will be deleted by the command.

#### Test accounts

The following accounts are defined in the fixture data and can be used for development and testing:

| E-mail                      | Password         | Role                        | Organisation |
| --------------------------- | ---------------- | --------------------------- | ------------ |
| admin@example.com           | myRandomPassword | _SystemAdministrator_       |              |
| admin+acmeinc@example.com   | myRandomPassword | _OrganisationAdministrator_ | Acme Inc.    |
| admin+othercorp@example.com | myRandomPassword | _OrganisationAdministrator_ | Other Corp   |

### 3. Start the server

```shell
cd backend
go run ./cli/ctl server --playground
```

The GraphQL API is now accessible at [http://localhost:8080/query](http://localhost:8080/query). A GraphQL playground to directly view the schema and execute queries / mutations can be called at [http://localhost:8080/](http://localhost:8080/) (if the `--playground` flag is set). See also [CLI](#cli) for a reference of all options.

!!! tip

    When developing, `refresh` can be used to automatically restart the backend server when files have been modified:

    ```shell
    go run ./cli/refresh
    ```

## Development

### Tests

All queries and mutations of the GraphQL API are covered by tests, which use stored fixed data (`test/fixtures`)
to capture the desired behavior with fixed data (`test/fixtures`).

The tests are mostly based on functional tests and test the different layers of the backend through the API.
This structure makes the tests independent of the actual implementation.
For the concurrent execution of the tests, schemas are created in the PostgreSQL database with random names and deleted when a test is
finished (`db.CreateTestDatabase`) - by this approach tests with complete DB access can be executed in parallel and isolated with high speed.

#### Execution of tests

!!! note

    Before running the tests, the [test database must be set up](#create-and-prepare-database-for-tests).

```shell
cd backend
go test ./...
```

### GraphQL API

The GraphQL API is implemented schema-first and is generated via [gqlgen](https://gqlgen.com/).

#### Changing the GraphQL schema

After changes to the GraphQL schema, the code must be regenerated. This is done via `go run ./cli/gqlgen`.

New resolvers are then available as a function and can be filled with the actual implementation.

### Persistence

*Models*

:    are based on Go structs in the `domain` package and are mapped to database fields with [construct](https://github.com/networkteam/construct) mappings.
     The models are built as DTOs for read operations. Construct also generates `...ChangeSet` structs that are used for type safe calls of write operations (`INSERT` / `UPDATE`).

*Repository*

:    operations are built as functions in the `persistence/repository` package and use constants and functions generated by `construct` as a base.
     Queries can be built directly using `*sql.DB` or [squirrel](https://github.com/Masterminds/squirrel) as query builder.
     Repository functions use `squirrel.BaseRunner` as interface for the current database connection or transaction.

*Queries*

:    use JSON results for complex data and Common Table Expressions (CTE) to build more complex queries.
     Necessary data is side-loaded into queries to avoid further selects (`N+1` problem).
     Eager loading of relations can be accomplished using `JSON_AGG()`.

#### Create a new migration

* Create a new file in `persistence/migrations` with the current date as prefix (`YYYMMDDHHmmss`)
* Assign a unique function name for up and down migration
* Migrations should be reversible

!!! info
    Migrations use [Goose](https://github.com/pressly/goose) and are embedded in the binary.

#### Update field mappings

After creating or modifying fields in the DB Models, `construct` must be called via `go generate ./persistence/repository`.
New models with construct mappings must be added to `mappings.go`.

## CLI

All backend functions are bundled in a CLI program `ctl`.
Configuration options are passed as arguments or environment variables.
The level of logging can be set via the `--verbosity` option and is set to `STDERR`.

!!! note

    Order of global and command options (see usage in help) matters.

### Reference

**ctl --help**

```
NAME:
   ctl - App CLI control

USAGE:
   ctl [global options] command [command options] [arguments...]

COMMANDS:
   server    Run the backend server
   migrate   Manage database migrations
   account   Manage accounts
   fixtures  Set up fixtures
   test      Test utilities
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbosity value, -v value  Verbosity: 0=fatal, 1=error, 2=warn, 3=info, 4=debug (default: 3) [$BACKEND_VERBOSITY]
   --postgres-dsn value         PostgreSQL connection DSN (default: "dbname=myproject-dev sslmode=disable") [$BACKEND_POSTGRES_DSN]
   --app-base-url value         Application base URL (default: "http://localhost:3000/") [$BACKEND_APP_BASE_URL]
   --smtp-host value            Host of SMTP for outgoing mails (default: "localhost") [$BACKEND_SMTP_HOST]
   --smtp-port value            SMTP Port for outgoing mails (default: 1025) [$BACKEND_SMTP_PORT]
   --smtp-user value            SMTP User for outgoing mails [$BACKEND_SMTP_USER]
   --smtp-password value        SMTP Password for outgoing mails [$BACKEND_SMTP_PASSWORD]
   --help, -h                   show help (default: false)
```

**ctl server --help**

```
NAME:
   ctl server - Run the backend server

USAGE:
   ctl server [command options] [arguments...]

OPTIONS:
   --address value             Listen on this address (default: "0.0.0.0:8080") [$BACKEND_ADDRESS]
   --playground                Enable GraphQL playground (default: false)
   --disable-ansi              Force disable ANSI log output and output log in logfmt format (default: false) [$BACKEND_DISABLE_ANSI]
   --sentry-dsn value          Sentry DSN (will be disabled if empty) [$SENTRY_DSN]
   --sentry-environment value  Sentry environment (default: "development") [$SENTRY_ENVIRONMENT]
   --sentry-release value      Release version for Sentry [$SENTRY_RELEASE]
   --help, -h                  show help (default: false)
```

**ctl migrate --help**

```
NAME:
   ctl migrate - Manage database migrations

USAGE:
   ctl migrate command [command options] [arguments...]

COMMANDS:
   up       Migrate up
   down     Migrate down
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

**ctl account --help**

```
NAME:
   ctl account - Manage accounts

USAGE:
   ctl account command [command options] [arguments...]

COMMANDS:
   create   Create account
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

**ctl account create --help**

```
NAME:
   ctl account create - Create account

USAGE:
   ctl account create [command options] [arguments...]

OPTIONS:
   --role value            (default: "SystemAdministrator")
   --email value
   --organisationId value
   --help, -h              show help (default: false)
```
