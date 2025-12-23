# Summary

`seldonio/python-builder` image is used as container for [`V1 Python Lint`](.github/workflows/python_lint.yml) and [`V1 Python Tests`](.github/workflows/python_tests.yml) Github Workflows.

Also this image is used to build python package and push it to PyPi, see `run_python_builder` command from the [root Makefile](Makefile).

# How to push new image

You could either use `push_to_registry` command from the [Makefile](./Makefile) or you could:
1. Run `gh workflow list` to get the id of the `Build & Push the Python Builder Image` workflow.
2. Run `gh workflow run <ID> -f docker-tag=<DOCKERTAG> --ref <BRANCH FROM WHICH YOU'D LIKE TO RUN THE WORKGLOW>`
