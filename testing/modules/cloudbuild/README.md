# KNFSD Cloud Build

This module configures Cloud Build with a private worker pool to support building and testing KNFSD proxy images and Terraform configuration.

## Inputs

* `project` - (Required) GCP Project ID to configure for building images.

* `region` - (Required) GCP Region that will be used to build images

* `network` - (Optional) Name of private VPC network to create for use by Cloud Build. Defaults to "knfsd-build".

* `worker_pool` - (Optional) Name of Cloud Build private worker pool to create. Defaults to "knfsd-build".

* `docker_repository` - (Optional) Name of the docker repository to create. This is used to store docker images used by Cloud Build. Defaults to "knfsd-docker".

## Outputs

* `network` - Full ID of the network for use with the `_NETWORK` substitution.

* `subnetwork` - Full ID of the subnetwork for use with the `_SUBNETWORK` substitution.

* `worker_pool` - Full ID of the worker pool for use with the `_WORKER_POOL` substitution.

* `docker_repository_url` - URL for the build docker repository for use with the `_DOCKER_REPOSITORY` substitution.

* `proxy_service_account` - Email address of the KNFSD proxy service account for use with the `_PROXY_SERVICE_ACCOUNT` substitution.
