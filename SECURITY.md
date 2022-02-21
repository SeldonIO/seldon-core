# Security Policy

This document provides an overview of the security policy of Seldon Core.

Seldon Core aims to follow the two following policies:

* Address CVEs in project dependencies by upgrading versions where possible
* Address CVEs in docker images by performing recommended upgrades

# Security Scans

As part of every release we perform a security scan. The scans include dependencies and docker image scans.

You can find the [exact commands that are used](https://github.com/SeldonIO/seldon-core/blob/master/.github/workflows/security_tests.yml) for the scans, together with the [reports generated](https://github.com/SeldonIO/seldon-core/actions/workflows/security_tests.yml) from each of these runs.

## Supported Versions

We use semver for our version management. We release security patches as a `patch version` for the latest maor.minor release.

## Reporting a Vulnerability

If you identify a vulnerability, if a public CVE the best way to report it is by opening an issue with the type "bug", the discussion can then take place on the ticket around next steps (ie updating library, reaching out to 3rd party projects, etc).

