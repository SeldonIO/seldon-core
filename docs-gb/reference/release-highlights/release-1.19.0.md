# Seldon Core Release 1.19.0 (Draft Release)

A summary of the main contributions to the [Seldon Core release 1.19.0](https://github.com/SeldonIO/seldon-core/releases/tag/v1.19.0).

### Overview
This release focuses on platform maintenance, security hardening, and ecosystem modernization. The main themes are extensive CVE remediation, a broad upgrade to Python 3.12 across images and tooling, dependency and base image refreshes, CI and workflow improvements, alongside documentation cleanup and migration work. No major new end-user features are introduced, but the release materially improves security posture, test coverage, and long-term maintainability.

### Security and CVE Fixes
- Multiple CVEs addressed across Go, Python, and container images, including the operator, executor, alibi-detect server, python-builder, and core builder images.
- Go runtime updated to a newer patch release to address known vulnerabilities.
- Autoscaling dependencies updated to remediate CVEs in the KEDA library.
- RClone upgraded and base images refreshed to address container-level CVEs.

### Dependency and Runtime Upgrades
- Python upgraded to Python 3.12 across:
  - Prepackaged servers
  - Alibi Detect and Alibi Explain servers
  - Python wrappers
- Conda base images migrated from Miniconda to Miniforge(OSS).
- UBI base images aligned and refreshed (UBI 9.7 where applicable).
- Protobuf, gRPC, Poetry, and other build-time dependencies updated.
- Kubernetes compatibility testing expanded to include Kubernetes 1.35 (1.23.x to 1.35.x).

### Images and Build System
- SBOM generation added to Docker image builds for improved license and supply-chain visibility.
- Deprecated GPU-related Docker images for the Python wrapper and both Alibi servers.

:warning: Note: Be aware of the changes to the [Bitnami public catalog](https://github.com/bitnami/charts/issues/35164) as they impact few examples in the documentation.

### Licensing and Compliance
Regeneration and cleanup of license metadata.

### Helm, Operator, and Configuration Changes
- Deprecated analytics Helm chart excluded from build workflows.
- Keda dependency upgraded to 2.17.3 and forward-compatible with 2.18.3(latest).

### Documentation
Documentation was migrated to GitBook. It can be found [here](https://docs.seldon.ai/seldon-core-1).

### Deprecations and Cleanup
- GPU-related files formally deprecated.
- Several examples and charts removed or disabled including:
    - "MLOps with Seldon and Jenkins Classic" notebook
    - "End-to-end MLOps with Seldon Core and Jenkins X" notebook
    - GPU examples and notebooks
    - Legacy Alibi Detect and Explain examples

### Backward Compatibility
- No intentional breaking API changes introduced in this release candidate.
- Users should validate custom images and notebooks against Python 3.12.
- Operators upgrading from older releases should review image and base-OS changes.
