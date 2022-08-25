# Rclone

We utilize [Rclone](https://rclone.org/) to copy model artifacts from a storage location to the model servers. This allows users to take advantage of Rclones support for over 40 cloud storage backends including Amazon S3, Google Storage and many others.

For local storage while developing see [here](../../getting-started/docker-installation/index.html#local-models).

For authorization needed for cloud storage when running on Kubernetes see [here](../../kubernetes/cloud-storage/index.html#kubernetes-secret).