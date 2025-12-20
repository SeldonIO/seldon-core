# Seldon Core 2 – Godog Test Architecture

This document describes the architecture of the BDD-style test suites for Seldon Core 2
using [godog](https://github.com/cucumber/godog) and Kubernetes.

The goals of this architecture are:

- Run **the same logical tests** against **different server configurations**.
- Drive tests via **feature tags + config**, not hard-coded setup.
- Provide a clean **domain-focused step API** (e.g. “models”, “pipelines”, …).
- Maintain an **in-memory view of Kubernetes resources** to make assertions easy and fast.
- Have flexibility to add future dependencies such as k6 or chaosmonkey

---

## 1. High-Level Overview

At a high level:

- **`TestMain`** creates and runs a `godog.TestSuite`.
- **`InitializeTestSuite`** creates long-lived test dependencies:
    - Kubernetes client(s)
    - A Kubernetes watcher for CRDs with `test-suite=godogs`
    - Reads/configures server setup from flags/config
    - Optionally deploys server replicas and other shared infra
- **`InitializeScenario`** runs per scenario:
    - Creates a fresh **World** object (per-scenario state holder)
    - Resets CRDs in the cluster (e.g. deletes test models)
    - Creates a fresh **Model** for the scenario
    - Registers domain-specific steps (e.g. model steps) against this World/Model
- **Feature files** (Gherkin) describe behavior in domain language:
    - e.g. “Given I have an "iris" model … Then the model should eventually become Ready”
- **A watcher** keeps an up-to-date in-memory store of CRDs with label `test-suite=godogs` for fast, poll-free
  assertions.

## Run a test case

```shell
  go test --godog.tags='@0' --godog.concurrency=1 -race
```

## List all steps
script to list out all the steps in the godog test suite, useful for creating tests with llms
```shell
  go run extract-steps/extract_steps.go  -root steps
```