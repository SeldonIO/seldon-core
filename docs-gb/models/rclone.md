---
description: >-
  Learn how to configure and use Rclone for model artifact storage in Seldon
  Core, including cloud storage integration and authentication.
---

# rClone

We utilize [Rclone](https://rclone.org/) to copy model artifacts from a storage location\
to the model servers. This allows users to take advantage of Rclones support for over 40\
cloud storage backends including Amazon S3, Google Storage and many others.

For authorization needed for cloud storage when running on Kubernetes see [here](../kubernetes/storage-secrets.md).
