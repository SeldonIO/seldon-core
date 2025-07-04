# Seldon Core Licensing

As of 22 January 2024, contributions to Seldon's open source and source available projects are governed by Seldon's Contributor License Agreement ("CLA").  For certain contributions previously contributed to Seldon's open source projects under the Apache 2.0 license, the individual contributions will remain open source under the Apache 2.0 license (see [here](https://github.com/SeldonIO/seldon-core/blob/v2.6.0/LICENSE)), but Seldon may include such contribution as part of a larger project that constitutes a derivative work of the project and is licensed as a whole under the Business Source License 1.1.  A list of individual contributions that are subject to the Apache 2.0 license are listed [here](oss_commits.md).

# Contributing to Seldon Core

_Before opening a pull request_ consider:

- Is the change important and ready enough to ask the community to spend time reviewing?
- Have you searched for existing, related issues and pull requests?
- Is the change being proposed clearly explained and motivated?

When you contribute code, you affirm that the contribution is your original work and that you
license the work to the project under the project's license. Whether or not you
state this explicitly, by submitting any copyrighted material via pull request, email, or
other means you agree to license the material under the project's license and
warrant that you have the legal authority to do so.

## PRs

Automated checks are run against PRs to ensure a certain level of quality.

One of these is a check that PR _titles_ conform to the [Conventional Commit](https://www.conventionalcommits.org/en/v1.0.0/) format.
This format ensures certain small but useful pieces of context are available.
Specifically, these are the _type_ of change being introduced, whether or not it is a breaking change, and an optional _scope_ of impact.
The permitted _types_ can be found in the [CI workflow](./.github/workflows/pr-title.yaml).
