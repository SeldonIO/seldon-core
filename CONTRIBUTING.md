# Contributing to Seldon Core

_Before opening a pull request_ consider:

- Is the change important and ready enough to ask the community to spend time reviewing?
- Have you searched for existing, related issues and pull requests?
- Is the change being proposed clearly explained and motivated?

When you contribute code, you affirm that the contribution is your original work and that you
license the work to the project under the project's open source license. Whether or not you
state this explicitly, by submitting any copyrighted material via pull request, email, or
other means you agree to license the material under the project's open source license and
warrant that you have the legal authority to do so.

## Coding conventions

We use [`pre-commit`](https://pre-commit.com/) to handle a number of Git hooks
which ensure that all changes to the codebase follow Seldon's code conventions.
It is recommended to set these up before making any change to the codebase.
Extra checks down the line will stop the build if the code is not compliant to
the style guide of each language in the repository.

To install it, follow the [official instructions](https://pre-commit.com/#install).
Once installed, run:

```console
$ pre-commit install
```

This will read the hooks defined in `.pre-commit-config.yaml` and install them
accordingly on your local repository.

### Java

To format our Java code we follow [Google's Java style
guide](https://google.github.io/styleguide/javaguide.html).
To make sure that the codebase remains consistent, we use
[`checkstyle`](https://github.com/checkstyle/checkstyle) as part of the `mvn validate` lifecycle.

To integrate these on your local editor, you can follow the official
instructions to [configure `checkstyle`
locally](https://checkstyle.org/beginning_development.html) and to [set-up
`google-java-format`](https://github.com/google/google-java-format#using-the-formatter).

### Python

To format our Python code we use [black](https://github.com/psf/black), the
heavily opinionated formatter.

To integrate it on your local editor, you can follow the official instructions
to [set-up black](https://github.com/psf/black#editor-integration).
