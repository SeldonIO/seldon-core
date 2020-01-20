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

We use [pre-commit](https://pre-commit.com/) to handle a number of Git hooks
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
[checkstyle](https://github.com/checkstyle/checkstyle) as part of the `mvn validate` lifecycle.

To integrate these on your local editor, you can follow the official
instructions to [configure checkstyle
locally](https://checkstyle.org/beginning_development.html) and to [set-up
google-java-format](https://github.com/google/google-java-format#using-the-formatter).

### Python

To format our Python code we use [black](https://github.com/psf/black), the
heavily opinionated formatter.

To integrate it on your local editor, you can follow the official instructions
to [set-up black](https://github.com/psf/black#editor-integration).

## Tests

Regardless of the package you are working on, we abstract the main tasks to a
`Makefile`.
Therefore, in order to run the tests, you should be able to just do:

```bash
$ make test
```

### Python

We use [pytest](https://docs.pytest.org/en/latest/) as our main test runner.
However, to ensure that tests run on the same version of the package that final
users will download from `pip` and pypi.org, we use
[tox](https://tox.readthedocs.io/en/latest/) on top of it.
To install both (plus other required plugins), just run:

```bash
$ make install_dev
```

Using `tox` we can run the entire test suite over different environments,
isolated between them.
You can see the different ones we currently use on the
[setup.cfg](https://github.com/SeldonIO/seldon-core/blob/master/python/setup.cfg)
file.
You can run your tests across all these environments using the standard `make test` [mentioned above](#Tests).
Alternatively, if you want to pass any extra parameters, you can also run `tox`
directly as:

```bash
$ tox
```

One of the caveats of `tox` is that, as the number of environments grows, so
does the time it takes to finish running the tests.
As a solution, during local development it may be recommended to run `pytest` directly
on your own environment.
You can do so as:

```bash
$ pytest
```

### Integration

As part of Seldon Core's test suite, we also run integration tests.
These spin up an actual Kubernetes cluster using
[Kind](https://github.com/kubernetes-sigs/kind) and deploy different
`SeldonDeployment` and resources.

You can read more about them and how to add new integration tests on [their
dedicated documentation](testing/scripts/README.md).
