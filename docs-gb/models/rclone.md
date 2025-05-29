---
description: Learn how to use Rclone in Seldon Core 2 for efficient model artifact management across 40+ cloud storage backends. This guide covers model artifact distribution, cloud storage integration with Amazon S3 and Google Storage, local development storage, and secure cloud storage authorization in Kubernetes environments.
---

# Rclone

We utilize [Rclone](https://rclone.org/) to copy model artifacts from a storage location
to the model servers. This allows users to take advantage of Rclones support for over 40
cloud storage backends including Amazon S3, Google Storage and many others.

For local storage while developing see [here](../getting-started/docker-installation.md#local-models).

For authorization needed for cloud storage when running on Kubernetes see [here](../../kubernetes/storage-secrets.md).
