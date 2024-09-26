# Go API backend boilerplate

This is a boilerplate for creating a GraphQL API backend with go.
It contains the setup for a monorepo application and is meant to be used as a starting point for creating a new project.
Because every project is different, all files can be freely changed to reflect the actual needs of the project.
E.g. it might use organisations for a multi-tenant application and authentication needs might differ.
That's why a boilerplate can be better shaped and has fewer abstractions than using a modular framework.

## Usage

* Checkout this Git repository
* Run `./create.sh` in the repository root to create a new project based on this boilerplate

**Example:**

```sh
./create.sh ../my-new-project com example myproject
```

The arguments will be used to replace placeholders for package names and other identifiers in the boilerplate files.

## Documentation

See [https://networkteam.github.io/go-apibackend-boilerplate/](docs) for more information about the code structure and development process.

The source for the documentation is part of the boilerplate and can be found in the `docs` directory.
