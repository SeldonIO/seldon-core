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

## Release notes

Our process to manage release notes is modelled after how the Kubernetes project handles them.
This process can be separated into 2 separate phases: 

- Adding notes on each PR. Happens at **PR creation time**.
- Compiling all PR notes before a release. Happens at **release time**.

### Adding notes on each PR

When a PR is created, a [Prow / Lighthouse
plugin](https://prow.k8s.io/command-help#release_note_none) will check if there
is a populated `release-note` block in the PR body such as:

````md
```release-note
Some public-facing release note.
```
````

If there isn't, the PR will be labelled as
`do-not-merge/release-note-label-needed`.
Note that to speed things up, the [default PR
template](https://github.com/SeldonIO/seldon-core/blob/master/.github/PULL_REQUEST_TEMPLATE.md)
will create an empty
`release-notes` block for you.
For PRs that don't need public-facing release notes (e.g. fixes on the
integration tests), you can use the `/release-note-none` Prow command.

#### Conventions

There are a number of conventions that we can use so that the changes are more
semantic.
These are mainly based around keywords which will affect how the release notes
will get displayed.

- Use the words `Added`, `Changed`, `Fixed`, `Removed` or `Deprecated` to
  describe the contents of the PR.
  For example:
  
  ````md
  ```release-note
  Added metadata support to Python wrapper
  ```
  ````

- Use the expression `Action required` to describe breaking changes.
  For example:

  ````md
  ```release-note
  Action required: The helm value `createResources` has been renamed
  `managerCreateResources`.
  ```
  ````

### Compiling all PR notes before a release

At release time, there is a [release-notes
command](https://github.com/kubernetes/release/blob/master/cmd/release-notes/README.md)
which crawls over all the PRs between 2 particular tags (e.g. `v1.1.0` to
`v1.2.0`), extracting the release-notes blocks.
These blocks can then be used to generate the final release notes.

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

### End to End Tests

As part of Seldon Core's test suite, we also run end to end tests.
These spin up an actual Kubernetes cluster using
[Kind](https://github.com/kubernetes-sigs/kind) and deploy different
`SeldonDeployment` and resources.

You can learn more about how to run them and how to add new test cases on
[their dedicated
documentation](https://docs.seldon.io/projects/seldon-core/en/latest/developer/e2e.html).
